package internal

import (
	"net"

	"authentication/internal/models"
	. "authentication/proto"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func ServeAsync(listener net.Listener, meta map[string]any, errch chan<- error) (shutdown func()) {
	var (
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

	go func() {
		var err error
		err = grpcSrv.Serve(listener)
		errch <- err
	}()

	return func() {
		_Logger.Warn("Server is shutting down")
		grpcSrv.GracefulStop()
		_Logger.Down()
	}
}
