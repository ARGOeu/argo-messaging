package config

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/spf13/viper"
)

// KafkaCfg holds kafka configuration
type KafkaCfg struct {
	Server string
	Topics []string
	Subs   map[string]string
}

// NewKafkaCfg creates a new kafka configuration object
func NewKafkaCfg(params ...string) *KafkaCfg {
	kcfg := KafkaCfg{}

	// If NewKafkaCfg is called with argument "LOAD" automatically load config
	for _, param := range params {
		if param == "LOAD" {
			kcfg.Load()
			return &kcfg
		}
	}

	return &kcfg
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
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - server", kcfg.Server)
	kcfg.Topics = viper.GetStringSlice("topics")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - topics", kcfg.Topics)
	kcfg.Subs = viper.GetStringMapString("subscriptions")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - subscriptions", kcfg.Subs)

}

// LoadStrJSON Loads configuration from a JSON string
func (kcfg *KafkaCfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewBuffer([]byte(input)))
	// Load Kafka configuration
	kcfg.Server = viper.GetString("server")
	kcfg.Topics = viper.GetStringSlice("topics")
	kcfg.Subs = viper.GetStringMapString("subscriptions")
}
