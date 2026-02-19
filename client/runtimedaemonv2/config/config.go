package config

import (
	_ "embed"
)

//go:embed config.yaml
var Data []byte
