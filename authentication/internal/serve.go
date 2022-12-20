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
	"github.com/d2jvkpn/go-web/pkg/orm"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// logger => database => tracing => grpc.Server => service registry
func ServeAsync(addr string, meta map[string]any, errch chan<- error) (err error) {

	var (
		enableOtel   bool
		enableConsul bool
		port         int
		listener     net.Listener
	)

	if _Relase {
		settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.InfoLevel, 256, nil)
	} else {
		settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.DebugLevel, 256, nil)
	}
	_Logger = settings.Logger.Named("server")

	// setup
	dsn := settings.Config.GetString("database.conn") + "/" +
		settings.Config.GetString("database.db")

	if _DB, err = models.Connect(dsn, !_Relase); err != nil {
		return err
	}

	enableOtel = settings.Config.GetBool("opentelemetry.enable")
	if enableOtel {
		if _CloseTracer, err = loadOtel(settings.Config); err != nil {
			return err
		}
	}

	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	_GrpcServer = grpc.NewServer(loadInterceptors(enableOtel)...)
	srv := models.NewServer()
	RegisterAuthServiceServer(_GrpcServer, srv)

	enableConsul = _ConsulClient != nil && _ConsulClient.Registry
	if enableConsul {
		if port, err = misc.PortFromAddr(addr); err != nil {
			return err
		}

		if err = _ConsulClient.GRPCRegister(port, false, _GrpcServer); err != nil {
			return err
		}
	}

	_Logger.Info(
		"Server is starting",
		zap.Bool("enableOtel", enableOtel),
		zap.Bool("enableConsul", enableConsul),
		zap.Any("meta", meta),
	)

	go func() {
		err := _GrpcServer.Serve(listener)
		errch <- err
	}()

	return nil
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

// service registry =>  grpc.Server => tracing => database => logger
func Shutdown() {
	var err error

	if _ConsulClient != nil && _ConsulClient.Registry {
		if err = _ConsulClient.Deregister(); err != nil {
			_Logger.Error(fmt.Sprintf("consul deregister: %v", err))
		} else {
			_Logger.Info("consul deregister")
		}
	}

	if _GrpcServer != nil {
		_Logger.Info("stop grpc server")
		_GrpcServer.GracefulStop()
	}

	if _CloseTracer != nil {
		_Logger.Info("close opentelemetry tracer")
		_CloseTracer()
	}

	if _DB != nil {
		if err = orm.CloseDB(_DB); err != nil {
			_Logger.Error(fmt.Sprintf("close database: %v", err))
		} else {
			_Logger.Info("close database")
		}
	}

	_Logger.Info("close logger")
	if settings.Logger != nil {
		settings.Logger.Down()
	}
}
