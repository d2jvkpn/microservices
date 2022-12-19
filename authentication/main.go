package main

import (
	// "fmt"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"authentication/internal"
)

func main() {
	var (
		release      bool
		addr, config string
		err          error
		shutdown     func()
	)

	flag.StringVar(&addr, "addr", ":20001", "grpc listening address")
	flag.StringVar(&config, "config", "configs/local.yaml", "configuration path")
	flag.BoolVar(&release, "release", false, "run in release mode")
	flag.Parse()

	if err = internal.Load(config, release); err != nil {
		log.Fatalln(err)
	}

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	log.Printf(">>> Greet RPC server: %q\n", addr)
	if shutdown, err = internal.ServeAsync(addr, release, errch); err != nil {
		log.Fatalln(err)
	}

	select {
	case err = <-errch:
	case <-quit:
		shutdown()
		err = <-errch
	}

	if err != nil {
		log.Fatalln(err)
	}
}
