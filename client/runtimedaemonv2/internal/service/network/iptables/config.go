package iptables

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Tables []struct {
		Name       string   `mapstructure:"table"`
		UserChains []string `mapstructure:"user_chains"`
		Rules      []string `mapstructure:"rules"`
	} `mapstructure:"iptables"`
}

func newConfig() (Config, error) {
	var result Config

	if err := viper.Unmarshal(&result); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return result, nil
}
