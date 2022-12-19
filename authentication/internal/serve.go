package internal

import (
	"net"

	"authentication/internal/models"
	. "authentication/proto"

	"google.golang.org/grpc"
)

func ServeAsync(addr string, release bool, errch chan<- error) (shutdown func(), err error) {
	var (
		listener    net.Listener
		grpcSrv     *grpc.Server
		interceptor *models.Interceptor
	)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	interceptor = models.NewInterceptor()
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

	return grpcSrv.GracefulStop, nil
}
