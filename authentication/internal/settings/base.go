package settings

import (
	// "fmt"
	"math/rand"
	"time"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/spf13/viper"
)

const (
	App = "authentication"
)

var (
	Logger *wrap.Logger
	Rng    *rand.Rand
	Config *viper.Viper
)

func init() {
	Rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}
