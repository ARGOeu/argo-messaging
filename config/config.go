package config

import (
	"bytes"
	"fmt"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/spf13/viper"
)

// KafkaCfg holds kafka configuration
type KafkaCfg struct {
	Server string
	Topics []string
}

// NewKafkaCfg creates a new kafka configuration object
func NewKafkaCfg() KafkaCfg {
	kcfg := KafkaCfg{}
	return kcfg
}

// Load the configuration
func (kcfg *KafkaCfg) Load() {

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/argo-messaging")
	viper.AddConfigPath(".")

	// Find and read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Errod trying to read the configuration file: %s \n", err))
	}

	// Load Kafka configuration
	kcfg.Server = viper.GetString("server")
	kcfg.Topics = viper.GetStringSlice("topics")
}

// LoadStrJSON Loads configuration from a JSON string
func (kcfg *KafkaCfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewBuffer([]byte(input)))
	// Load Kafka configuration
	kcfg.Server = viper.GetString("server")
	kcfg.Topics = viper.GetStringSlice("topics")
}
