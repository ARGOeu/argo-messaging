package config

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/spf13/viper"
)

// APICfg holds kafka configuration
type APICfg struct {
	// values
	BrokerHost string
	StoreHost  string
	StoreDB    string
}

// NewAPICfg creates a new kafka configuration object
func NewAPICfg(params ...string) *APICfg {
	cfg := APICfg{}

	// If NewKafkaCfg is called with argument "LOAD" automatically load config
	for _, param := range params {
		if param == "LOAD" {
			cfg.Load()
			return &cfg
		}
	}

	return &cfg
}

// Load the configuration
func (cfg *APICfg) Load() {

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/argo-messaging")
	viper.AddConfigPath(".")

	// Find and read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Errod trying to read the configuration file: %s \n", err))
	}

	// Load Kafka configuration
	cfg.BrokerHost = viper.GetString("broker_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - broker_host", cfg.BrokerHost)
	cfg.StoreHost = viper.GetString("store_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_host", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_db", cfg.StoreDB)

}

// LoadStrJSON Loads configuration from a JSON string
func (cfg *APICfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewBuffer([]byte(input)))
	// Load Kafka configuration
	cfg.BrokerHost = viper.GetString("broker_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - broker_host", cfg.BrokerHost)
	cfg.StoreHost = viper.GetString("store_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_host", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_db", cfg.StoreDB)

}
