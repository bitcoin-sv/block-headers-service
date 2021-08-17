package config

import (
	"github.com/spf13/viper"
)

func SetDefaults() {
	viper.SetDefault(EnvHeaderType, "node")

	// Node defaults
	viper.SetDefault(EnvNodeHost, "localhost")
	viper.SetDefault(EnvNodePort, 18332)
	viper.SetDefault(EnvNodeUser, "bitcoin")
	viper.SetDefault(EnvNodePassword, "bitcoin")
}
