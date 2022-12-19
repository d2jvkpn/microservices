package internal

import (
	// "fmt"

	"authentication/internal/models"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
)

func Load(config string, release bool) (err error) {
	var vc *viper.Viper

	vc = viper.New()
	vc.SetConfigName(App)
	vc.SetConfigFile(config)
	if err = vc.ReadInConfig(); err != nil {
		return err
	}

	dsn := vc.GetString("database.conn") + "/" + vc.GetString("database.db")
	if _, err = models.Connect(dsn, !release); err != nil {
		return err
	}

	return nil
}

func LoadWithConsul(consul string, release bool) (err error) {
	var (
		_ConsulClient *wrap.ConsulClient
		vc            *viper.Viper
	)

	if _ConsulClient, err = wrap.NewConsulClient(consul, "consul"); err != nil {
		return err
	}

	if vc, err = _ConsulClient.GetKV(App); err != nil {
		return err
	}

	dsn := vc.GetString("database.conn") + "/" + vc.GetString("database.db")
	if _, err = models.Connect(dsn, !release); err != nil {
		return err
	}

	return nil
}
