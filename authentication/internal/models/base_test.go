package models

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/d2jvkpn/go-web/pkg/misc"
	"github.com/spf13/viper"
)

var (
	testAddr string          = "127.0.0.1:20001"
	testFlag *flag.FlagSet   = nil
	testCtx  context.Context = context.Background()
)

func TestMain(m *testing.M) {
	var (
		config string
		err    error
		vc     *viper.Viper
	)

	testFlag = flag.NewFlagSet("testFlag", flag.ExitOnError)
	flag.Parse() // must do

	testFlag.StringVar(&config, "config", "configs/local.yaml", "config filepath")

	testFlag.Parse(flag.Args())
	fmt.Printf("~~~ load config %s\n", config)

	defer func() {
		if err != nil {
			fmt.Printf("!!! TestMain: %v\n", err)
			os.Exit(1)
		}
	}()

	if config, err = misc.RootFile(config); err != nil {
		return
	}

	vc = viper.New()
	vc.SetConfigName("test config")
	vc.SetConfigFile(config)

	if err = vc.ReadInConfig(); err != nil {
		return
	}

	dsn := vc.GetString("database.conn") + "/" + vc.GetString("database.db")
	if _DB, err = Connect(dsn, true); err != nil {
		return
	}

	m.Run()
}
