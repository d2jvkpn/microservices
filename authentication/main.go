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

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	var (
		release      bool
		addr, config string
		err          error
		vc           *viper.Viper
		listener     net.Listener
		grpcSrv      *grpc.Server
	)

	flag.StringVar(&addr, "addr", ":20001", "grpc listening address")
	flag.StringVar(&config, "config", "configs/local.yaml", "configuration path")
	flag.BoolVar(&release, "release", false, "run in release mode")
	flag.Parse()

	vc = viper.New()
	vc.SetConfigName("authentication config")
	vc.SetConfigFile(config)
	if err = vc.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}

	dsn := vc.GetString("database.conn") + "/" + vc.GetString("database.db")
	if _, err = models.Connect(dsn, !release); err != nil {
		log.Fatalln(err)
	}

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
