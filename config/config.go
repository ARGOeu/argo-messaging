package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// kafka Configuration
type configKafka struct {
	Server string
	Topics []string
}

// Global Kafka configuration
var Kafka configKafka

// Load the configuration
func Load() {

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/argo-messaging")
	viper.AddConfigPath(".")

	// Find and read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Errod trying to read the configuration file: %s \n", err))
	}

	// Load Kafka configuration
	Kafka.Server = viper.GetString("server")
	Kafka.Topics = viper.GetStringSlice("topics")
}
