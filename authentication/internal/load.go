package internal

import (
	// "fmt"

	"authentication/internal/models"

	"github.com/spf13/viper"
)

func Load(config string, release bool) (err error) {
	var vc *viper.Viper

	vc = viper.New()
	vc.SetConfigName("authentication config")
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
