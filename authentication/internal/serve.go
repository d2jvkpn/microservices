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
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func ServeAsync(addr string, meta map[string]any, errch chan<- error) (
	shutdown func(), err error) {

	var (
		enableOtel  bool
		port        int
		listener    net.Listener
		grpcSrv     *grpc.Server
		interceptor *models.Interceptor
		closeTracer func()
	)

	settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.InfoLevel, 256, nil)
	_Logger = settings.Logger.Named("server")
	_Logger.Info("Server is starting", zap.Any("meta", meta))

	enableOtel = settings.Config.GetBool("opentelemetry.enable")
	interceptor = models.NewInterceptor(_Logger.Named("grpc"))

	options := make([]grpc.ServerOption, 0, 4)
	if enableOtel {
		options = append(
			options,
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		)
	}
	options = append(options, interceptor.Unary(), interceptor.Stream())

	grpcSrv = grpc.NewServer(options...)

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
		if closeTracer, err = loadOtel(); err != nil {
			return nil, err
		}
	}

	go func() {
		var err error
		err = grpcSrv.Serve(listener)
		errch <- err
	}()

	shutdown = func() {
		_Logger.Warn("Server is shutting down")
		grpcSrv.GracefulStop()
		if closeTracer != nil {
			closeTracer()
		}
	}

	return shutdown, nil
}

func loadOtel() (closeTracer func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	str := settings.Config.GetString("opentelemetry.address")
	secure := settings.Config.GetBool("opentelemetry.secure")

	closeTracer, err = cloud_native.LoadTracer(ctx, str, settings.App, secure)
	if err != nil {
		return nil, fmt.Errorf("cloud_native.LoadTracer: %s, %w", str, err)
	}

	return closeTracer, nil
}
