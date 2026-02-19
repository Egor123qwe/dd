package network

import "github.com/spf13/viper"

type config struct {
	loginServer string
}

func newConfig() config {
	return config{
		loginServer: viper.GetString("network.tailscale.login_server"),
	}
}
