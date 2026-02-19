package shared

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	RcloneConfigDirPath = "/.config/rclone"

	RcloneConfigFile = "rclone.conf"
)

type config struct {
	storageURL string
	rclonePath string
}

func newConfig() config {
	return config{
		storageURL: viper.GetString("docker.sharedFolder.storageAPI"),
		rclonePath: viper.GetString("docker.sharedFolder.rclonePath"),
	}
}

func initConfigRcloneDir() (string, error) {
	userDir, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to getting home directory %w", err)
	}

	configDirPath := filepath.Join(userDir.HomeDir, RcloneConfigDirPath)

	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return "", fmt.Errorf("falid to init storage: %w", err)
	}

	configPath := filepath.Join(configDirPath, RcloneConfigFile)

	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("falid to init storage: %w", err)
	}

	file.Close()

	return configPath, nil
}
