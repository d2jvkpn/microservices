package internal

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/d2jvkpn/microservices/authentication/internal/models"
	"github.com/d2jvkpn/microservices/authentication/internal/settings"
	. "github.com/d2jvkpn/microservices/authentication/proto"

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

// database => tracing => grpc.Server => service registry => logger
func ServeAsync(addr string, meta map[string]any, errch chan<- error) (err error) {
	var (
		enableOtel   bool
		enableConsul bool
		port         int
		listener     net.Listener
	)

	// database
	dsn := settings.Config.GetString("postgres.conn") + "/" +
		settings.Config.GetString("postgres.db")

	if _DB, err = models.Connect(dsn, !settings.Release); err != nil {
		return err
	}

	// opentelemetry
	enableOtel = settings.Config.GetBool("opentelemetry.enable")
	if enableOtel {
		if _CloseTracer, err = setupOtel(settings.Config); err != nil {
			return err
		}
	}

	// grpc
	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}

	_GrpcServer = grpc.NewServer(setupInterceptors(enableOtel)...)
	srv := models.NewServer()
	RegisterAuthServiceServer(_GrpcServer, srv)

	// consul
	enableConsul = _ConsulClient != nil && _ConsulClient.Registry
	if enableConsul {
		if port, err = misc.PortFromAddr(addr); err != nil {
			return err
		}

		if err = _ConsulClient.GRPCRegister(port, false, _GrpcServer); err != nil {
			return err
		}
	}

	// logger
	if settings.Release {
		settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.InfoLevel, 256, nil)
	} else {
		settings.Logger = wrap.NewLogger("logs/authentication.log", zapcore.DebugLevel, 256, nil)
	}
	_Logger = settings.Logger.Named("server")

	_Logger.Info(
		"Server is starting",
		zap.Bool("enableOtel", enableOtel),
		zap.Bool("enableConsul", enableConsul),
		zap.Any("meta", meta),
	)

	// serve
	go func() {
		err := _GrpcServer.Serve(listener)
		errch <- err
	}()

	return nil
}

func setupInterceptors(enableOtel bool) []grpc.ServerOption {
	interceptor := models.NewInterceptor()

	uIntes := make([]grpc.UnaryServerInterceptor, 0, 2)
	sIntes := make([]grpc.StreamServerInterceptor, 0, 2)

	if enableOtel {
		uIntes = append(uIntes, otelgrpc.UnaryServerInterceptor( /*opts ...Option*/ ))
		sIntes = append(sIntes, otelgrpc.StreamServerInterceptor( /*opts ...Option*/ ))
	}
	// grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	// grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	uIntes = append(uIntes, interceptor.Unary())
	sIntes = append(sIntes, interceptor.Stream())

	uInte := grpc_middleware.ChainUnaryServer(uIntes...)
	sInte := grpc_middleware.ChainStreamServer(sIntes...)
	return []grpc.ServerOption{grpc.UnaryInterceptor(uInte), grpc.StreamInterceptor(sInte)}
}

func setupOtel(vc *viper.Viper) (closeTracer func(), err error) {
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

// service registry => grpc.Server => tracing => database => logger
func Shutdown() {
	var err error

	errorf := func(format string, a ...any) {
		if _Logger == nil {
			return
		}
		_Logger.Error(fmt.Sprintf(format, a...))
	}

	infof := func(format string, a ...any) {
		if _Logger == nil {
			return
		}
		_Logger.Info(fmt.Sprintf(format, a...))
	}

	if _ConsulClient != nil && _ConsulClient.Registry {
		if err = _ConsulClient.Deregister(); err != nil {
			errorf("consul deregister: %v", err)
		} else {
			infof("consul deregister")
		}
	}

	if _GrpcServer != nil {
		infof("stop grpc server")
		_GrpcServer.GracefulStop()
	}

	if _CloseTracer != nil {
		infof("close tracer")
		_CloseTracer()
	}

	if _DB != nil {
		if err = orm.CloseDB(_DB); err != nil {
			errorf(fmt.Sprintf("close database: %v", err))
		} else {
			infof("close database")
		}
	}

	infof("close logger")
	if settings.Logger != nil {
		settings.Logger.Down()
	}
}
