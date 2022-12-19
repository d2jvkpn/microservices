package settings

import (
	// "fmt"
	"log"
	"math/rand"
	"time"

	"github.com/d2jvkpn/go-web/pkg/orm"
	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	App = "authentication"
)

var (
	Logger       *wrap.Logger
	Rng          *rand.Rand
	Config       *viper.Viper
	DB           *gorm.DB
	ConsulClient *wrap.ConsulClient
)

func init() {
	Rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Shutdown() {
	var err error

	if ConsulClient != nil && ConsulClient.Registry {
		if e := ConsulClient.Deregister(); e != nil {
			log.Printf("consul deregister: %v\n", err)
		}
	}

	if DB != nil {
		if err = orm.CloseDB(DB); err != nil {
			log.Printf("close database: %v\n", err)
		}
	}

	if Logger != nil {
		Logger.Down()
	}
}
