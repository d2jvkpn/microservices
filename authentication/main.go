package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"authentication/internal"
	"authentication/internal/settings"

	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
)

var (
	//go:embed project.yaml
	_Project []byte
)

func init() {
	misc.RegisterLogPrinter()
}

func main() {
	var (
		release      bool
		addr, config string
		consul       string
		err          error
		project      *viper.Viper
		shutdown     func()
	)

	if project, err = wrap.ConfigFromBytes(_Project, "yaml"); err != nil {
		log.Fatalln(err)
	}
	meta := misc.BuildInfo()
	meta["project"] = project.GetString("project")
	meta["version"] = project.GetString("version")

	flag.StringVar(&addr, "addr", ":20001", "grpc listening address")
	flag.StringVar(&config, "config", "configs/local.yaml", "configuration path")
	flag.StringVar(
		&consul, "consul", "",
		"consul config path, set -config to empty if you needs to read config from consul",
	)
	flag.BoolVar(&release, "release", false, "run in release mode")

	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(
			output, "%s\n\nUsage of %s:\n",
			misc.BuildInfoText(meta), filepath.Base(os.Args[0]),
		)

		flag.PrintDefaults()

		fmt.Fprintf(output, "\nConfig template:\n```yaml\n%s```\n", project.GetString("config"))
		fmt.Fprintf(output, "\nConsul template:\n```yaml\n%s```\n", wrap.ConsulConfigDemo())
	}
	flag.Parse()

	meta["-config"] = config
	meta["-addr"] = addr
	meta["-release"] = release
	meta["pid"] = os.Getpid()

	if err = internal.Load(config, consul, release); err != nil {
		log.Fatalln(err)
	}

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)

	log.Printf(">>> Greet RPC server: %q\n", addr)
	if shutdown, err = internal.ServeAsync(addr, meta, errch); err != nil {
		settings.Shutdown()
		log.Fatalln(err)
	}

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case err = <-errch:
	case <-quit:
		shutdown()
		err = <-errch
	}

	if err != nil {
		settings.Shutdown()
		log.Fatalln(err)
	} else {
		log.Println("<<< Exit")
	}
}
