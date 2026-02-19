package piko

import "github.com/spf13/viper"

type config struct {
	linkTemplate string
	version      string
	serverURL    string
}

func newConfig() config {
	return config{
		linkTemplate: viper.GetString("network.piko.link_template"),
		version:      viper.GetString("network.piko.version"),
		serverURL:    viper.GetString("network.piko.server_url"),
	}
}
