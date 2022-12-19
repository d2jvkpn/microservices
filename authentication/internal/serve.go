package internal

import (
	"net"

	"authentication/internal/models"
	. "authentication/proto"

	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func ServeAsync(addr string, meta map[string]any, errch chan<- error) (
	shutdown func(), err error) {
	var (
		port        int
		listener    net.Listener
		grpcSrv     *grpc.Server
		interceptor *models.Interceptor
	)

	_Logger = wrap.NewLogger("logs/authentication.log", zapcore.InfoLevel, 256, nil)
	_Logger.Info("Server is starting", zap.Any("meta", meta))

	interceptor = models.NewInterceptor(_Logger.Named("grpc"))
	grpcSrv = grpc.NewServer(
		interceptor.Unary(),
		interceptor.Stream(),
	)

	srv := models.NewServer()
	RegisterAuthServiceServer(grpcSrv, srv)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	consulEnabled := _ConsulClient != nil && _ConsulClient.Registry

	if consulEnabled {
		if port, err = misc.PortFromAddr(addr); err != nil {
			return nil, err
		}

		if err = _ConsulClient.GRPCRegister(port, false, grpcSrv); err != nil {
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

		if consulEnabled {
			if e := _ConsulClient.Deregister(); e != nil {
				_Logger.Error("consul deregister", zap.Any("error", e))
			}
		}

		_Logger.Down()
	}

	return shutdown, nil
}
