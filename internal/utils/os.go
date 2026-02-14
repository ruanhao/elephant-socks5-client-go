package utils

import (
	"os"

	"github.com/ruanhao/elephant/internal/config"
)

func GetOsHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return config.AppConfig.Alias
	}
	return hostname
}
