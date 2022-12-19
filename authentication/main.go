package main

import (
	// "fmt"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"authentication/internal/models"
	. "authentication/proto"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	// "google.golang.org/grpc/status"
)

func main() {
	var (
		addr     string
		err      error
		listener net.Listener
		grpcSrv  *grpc.Server
	)

	flag.StringVar(&addr, "addr", ":20001", "grpc listening address")
	flag.Parse()

	if listener, err = net.Listen("tcp", addr); err != nil {
		log.Fatalln(err)
	}

	grpcSrv = grpc.NewServer(
	// grpc.UnaryInterceptor(unaryServerInterceptor),
	// grpc.StreamInterceptor(streamServerInterceptor),
	)

	srv := models.NewServer()
	RegisterAuthServiceServer(grpcSrv, srv)

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf(">>> Greet RPC server: %q\n", addr)
		errch <- grpcSrv.Serve(listener)
	}()

	select {
	case err = <-errch:
	case <-quit:
		grpcSrv.GracefulStop()
		err = <-errch
	}

	if err != nil {
		log.Fatalln(err)
	}
}
