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
	flag.BoolVar(&release, "release", false, "run in release mode")

	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(
			output, "%s\n\nUsage of %s:\n",
			misc.BuildInfoText(meta), filepath.Base(os.Args[0]),
		)

		flag.PrintDefaults()

		fmt.Fprintf(output, "\nConfig template:\n```yaml\n%s```\n", project.GetString("config"))
	}
	flag.Parse()

	meta["-config"] = config
	meta["-addr"] = addr
	meta["-release"] = release
	meta["pid"] = os.Getpid()

	if err = internal.Load(config, release); err != nil {
		log.Fatalln(err)
	}

	errch, quit := make(chan error, 1), make(chan os.Signal, 1)

	log.Printf(">>> Greet RPC server: %q\n", addr)
	if shutdown, err = internal.ServeAsync(addr, meta, errch); err != nil {
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
		log.Fatalln(err)
	}
}
