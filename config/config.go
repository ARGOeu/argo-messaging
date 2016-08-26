package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
)

// APICfg holds kafka configuration
type APICfg struct {
	// values
	BindIP    string
	Port      int
	ZooHosts  []string
	StoreHost string
	StoreDB   string
	Cert      string
	CertKey   string
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

type brokerInfo struct {
	Host string
	Port int
}

// GetZooList gets list from zookeeper
func (cfg *APICfg) GetZooList() []string {
	zConn, _, err := zk.Connect(cfg.ZooHosts, time.Second)
	if err != nil {
		panic(err)
	}
	brIDs, _, err := zConn.Children("/brokers/ids")
	if err != nil {
		panic(err)
	}

	peerList := []string{}

	for brID := range brIDs {
		data, _, err := zConn.Get("/brokers/ids/" + strconv.Itoa(brID))
		if err != nil {
			panic(err)
		}
		var brk brokerInfo
		json.Unmarshal(data, &brk)
		peer := brk.Host + ":" + strconv.Itoa(brk.Port)
		peerList = append(peerList, peer)

	}

	return peerList
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
	cfg.BindIP = viper.GetString("bind_ip")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - bind_ip", cfg.BindIP)
	cfg.Port = viper.GetInt("port")
	log.Printf("%s\t%s\t%s:%d", "INFO", "CONFIG", "Parameter Loaded - port", cfg.Port)
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - zookeeper_hosts", cfg.ZooHosts)
	cfg.StoreHost = viper.GetString("store_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_host", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_db", cfg.StoreDB)
	cfg.Cert = viper.GetString("certificate")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - certificate", cfg.Cert)
	cfg.CertKey = viper.GetString("certificate_key")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - certificate_key", cfg.CertKey)

}

// LoadStrJSON Loads configuration from a JSON string
func (cfg *APICfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewBuffer([]byte(input)))
	// Load Kafka configuration
	cfg.BindIP = viper.GetString("bind_ip")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - bind_ip", cfg.BindIP)
	cfg.Port = viper.GetInt("port")
	log.Printf("%s\t%s\t%s:%d", "INFO", "CONFIG", "Parameter Loaded - port", cfg.Port)
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - zookeeper_hosts", cfg.ZooHosts)
	cfg.StoreHost = viper.GetString("store_host")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_host", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - store_db", cfg.StoreDB)
	cfg.Cert = viper.GetString("certificate")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - certificate", cfg.Cert)
	cfg.CertKey = viper.GetString("certificate_key")
	log.Printf("%s\t%s\t%s:%s", "INFO", "CONFIG", "Parameter Loaded - certificate_key", cfg.CertKey)

}
