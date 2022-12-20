package internal

import (
	"fmt"

	"authentication/internal/settings"

	"github.com/d2jvkpn/go-web/pkg/wrap"
)

func Load(config string, consul string, release bool) (err error) {
	if config == "" && consul == "" {
		return fmt.Errorf("both config and  consul are empty")
	}

	if consul != "" {
		if _ConsulClient, err = wrap.NewConsulClient(consul, "consul"); err != nil {
			return err
		}
	}

	if config == "" {
		settings.Config, err = _ConsulClient.GetKV(settings.App)
	} else {
		settings.Config, err = wrap.OpenConfig(config)
	}
	if err != nil {
		return err
	}

	settings.Release = release

	return nil
}
