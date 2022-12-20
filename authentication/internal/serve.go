package internal

import (
	"context"
	"fmt"
	"net"
	"time"

	"authentication/internal/models"
	"authentication/internal/settings"
	. "authentication/proto"

	"github.com/d2jvkpn/go-web/pkg/cloud_native"
	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func ServeAsync(addr string, meta map[string]any, errch chan<- error) (
	shutdown func(), err error) {

	var (
		enableOtel  bool
		port        int
		listener    net.Listener
		grpcSrv     *grpc.Server
		closeTracer func()
	)

	settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.InfoLevel, 256, nil)
	_Logger = settings.Logger.Named("server")

	enableOtel = settings.Config.GetBool("opentelemetry.enable")
	grpcSrv = grpc.NewServer(loadInterceptors(enableOtel)...)

	srv := models.NewServer()
	RegisterAuthServiceServer(grpcSrv, srv)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	consulEnabled := settings.ConsulClient != nil && settings.ConsulClient.Registry
	if consulEnabled {
		if port, err = misc.PortFromAddr(addr); err != nil {
			return nil, err
		}

		if err = settings.ConsulClient.GRPCRegister(port, false, grpcSrv); err != nil {
			return nil, err
		}
	}

	// setup
	dsn := settings.Config.GetString("database.conn") + "/" +
		settings.Config.GetString("database.db")

	if settings.DB, err = models.Connect(dsn, !_Relase); err != nil {
		return nil, err
	}
	// TODO: ?? close connection with database

	if enableOtel {
		if closeTracer, err = loadOtel(settings.Config); err != nil {
			return nil, err
		}
	}

	_Logger.Info("Server is starting", zap.Any("meta", meta), zap.Bool("enableOtel", enableOtel))
	go func() {
		var err error
		err = grpcSrv.Serve(listener)
		errch <- err
	}()

	shutdown = func() {
		_Logger.Warn("Server is shutting down")
		grpcSrv.GracefulStop()
		if enableOtel { // closeTracer != nil
			closeTracer()
		}
	}

	return shutdown, nil
}

func loadInterceptors(enableOtel bool) []grpc.ServerOption {
	interceptor := models.NewInterceptor(_Logger.Named("grpc"))

	uIntes := make([]grpc.UnaryServerInterceptor, 0, 2)
	sIntes := make([]grpc.StreamServerInterceptor, 0, 2)

	if enableOtel {
		uIntes = append(uIntes, otelgrpc.UnaryServerInterceptor())
		sIntes = append(sIntes, otelgrpc.StreamServerInterceptor())
	}
	// grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	// grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	uIntes = append(uIntes, interceptor.Unary())
	sIntes = append(sIntes, interceptor.Stream())

	uInte := grpc_middleware.ChainUnaryServer(uIntes...)
	sInte := grpc_middleware.ChainStreamServer(sIntes...)
	return []grpc.ServerOption{grpc.UnaryInterceptor(uInte), grpc.StreamInterceptor(sInte)}
}

func loadOtel(vc *viper.Viper) (closeTracer func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	str := vc.GetString("opentelemetry.address")
	secure := vc.GetBool("opentelemetry.secure")

	closeTracer, err = cloud_native.LoadTracer(ctx, str, settings.App, secure)
	if err != nil {
		return nil, fmt.Errorf("cloud_native.LoadTracer: %s, %w", str, err)
	}

	return closeTracer, nil
}
