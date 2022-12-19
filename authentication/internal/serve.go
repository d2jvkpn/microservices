package internal

import (
	"net"

	"authentication/internal/models"
	. "authentication/proto"

	"google.golang.org/grpc"
)

func ServeAsync(addr string, release bool, errch chan<- error) (shutdown func(), err error) {
	var (
		listener net.Listener
		grpcSrv  *grpc.Server
	)

	if listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	grpcSrv = grpc.NewServer(
	// grpc.UnaryInterceptor(unaryServerInterceptor),
	// grpc.StreamInterceptor(streamServerInterceptor),
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
