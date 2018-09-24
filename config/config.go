package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// APICfg holds kafka configuration
type APICfg struct {
	// values
	BindIP       string
	Port         int
	ZooHosts     []string
	KafkaZnode   string //The Zookeeper znode used by Kafka
	StoreHost    string
	StoreDB      string
	Cert         string
	CertKey      string
	ResAuth      bool
	ServiceToken string
	LogLevel     string
	PushEnabled  bool
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
		if param == "LOADTEST" {
			cfg.LoadTest()
		}
	}

	return &cfg
}

type brokerInfo struct {
	Host string
	Port int
}

// GetBrokerInfo is a wrapper over GetZooList which retrieves broker information from zookeeper
func (cfg *APICfg) GetBrokerInfo() []string {
	// Iterate trying to retrieve broker information from zookeeper
	for {
		brkList, err := cfg.GetZooList()
		if err != nil {
			// If error retrieving info try again in 3 seconds
			time.Sleep(3 * time.Second)
			log.Error("ZK", "\t", "Broker list invalid: ", err.Error())
		} else {
			// Info retrieved succesfully so continue
			log.Info("ZK", "\t", "Discovered broker list:", brkList)
			return brkList
		}

	}
}

// GetZooList gets broker list from zookeeper
func (cfg *APICfg) GetZooList() ([]string, error) {
	peerList := []string{}
	log.Info("ZK", "\t", "Trying to connect zookeper hosts: ", cfg.ZooHosts, " ...")
	zConn, _, err := zk.Connect(cfg.ZooHosts, time.Second)
	// Check if indeed connected and can read
	_, _, _, err = zConn.ChildrenW("/")
	if err != nil {
		zConn.Close()
		return peerList, err
	}

	log.Info("ZK", "\t", "Connected to zookeper hosts: ", cfg.ZooHosts)
	log.Info("ZK", "\t", "Attempting to read broker information")
	brIDs, _, err := zConn.Children(cfg.KafkaZnode + "/brokers/ids")
	if err != nil {
		return peerList, err
	}

	for _, brID := range brIDs {
		data, _, err := zConn.Get(cfg.KafkaZnode + "/brokers/ids/" + brID)

		if err != nil {
			return peerList, err
		}

		var brk brokerInfo
		json.Unmarshal(data, &brk)
		peer := brk.Host + ":" + strconv.Itoa(brk.Port)
		peerList = append(peerList, peer)

	}

	if len(peerList) == 0 {
		return peerList, errors.New("empty")
	}

	return peerList, nil
}

func setLogLevel(logLvl string) {

	switch logLvl {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		break
	case "INFO":
		log.SetLevel(log.InfoLevel)
		break
	case "WARNING":
		log.SetLevel(log.WarnLevel)
		break
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
		break
	case "FATAL":
		log.SetLevel(log.FatalLevel)
		break
	default:
		log.SetLevel(log.InfoLevel)
	}

}

// LoadTest the configuration
func (cfg *APICfg) LoadTest() {

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/argo-messaging")
	viper.AddConfigPath(".")

	// Find and read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Errod trying to read the configuration file: %s \n", err))
	}

	// Load Kafka configuration

	// First check log level parameter and set logger
	cfg.LogLevel = viper.GetString("log_level")
	setLogLevel(cfg.LogLevel)
	log.Info("CONFIG", "\t", "Parameter Loaded - log_level: ", cfg.LogLevel)
	// Then load rest of the parameters

	cfg.BindIP = viper.GetString("bind_ip")
	log.Info("CONFIG", "\t", "Parameter Loaded - bind_ip: ", cfg.BindIP)
	cfg.Port = viper.GetInt("port")
	log.Info("CONFIG", "\t", "Parameter Loaded - port: ", cfg.Port)
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.Info("CONFIG", "\t", "Parameter Loaded - zookeeper_hosts: ", cfg.ZooHosts)
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.Info("CONFIG", "\t", "Parameter Loaded - kafka_znode: ", cfg.KafkaZnode)
	cfg.StoreHost = viper.GetString("store_host")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_host: ", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_db: ", cfg.StoreDB)
	cfg.Cert = viper.GetString("certificate")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate: ", cfg.Cert)
	cfg.CertKey = viper.GetString("certificate_key")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate_key: ", cfg.CertKey)
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.Info("CONFIG", "\t", "Parameter Loaded - per_resource_auth: ", cfg.CertKey)
	cfg.ServiceToken = viper.GetString("service_token")
	log.Info("CONFIG", "\t", "Parameter Loaded - service_token: ", cfg.ServiceToken)
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.Info("CONFIG", "\t", "Parameter Loaded - push_enabled: ", cfg.PushEnabled)

}

// Load the configuration
func (cfg *APICfg) Load() {
	// Set Flags
	var configPath *string

	if pflag.Parsed() == false {

		pflag.String("log-level", "INFO", "set the desired log level")
		viper.BindPFlag("log_level", pflag.Lookup("log-level"))

		pflag.String("bind-ip", "localhost", "ip address to listen to")
		viper.BindPFlag("bind_ip", pflag.Lookup("bind-ip"))

		pflag.Int("port", 8080, "port number to listen to")
		viper.BindPFlag("port", pflag.Lookup("port"))

		pflag.StringSlice("zookeeper-hosts", []string{"localhost"}, "list of zookeeper hosts to connect to")
		viper.BindPFlag("zookeeper_hosts", pflag.Lookup("zookeeper-hosts"))

		pflag.String("kafka-znode", "", "kafka zookeeper node name")
		viper.BindPFlag("kafka_znode", pflag.Lookup("kafka-znode"))

		pflag.String("store-host", "localhost", "datastore (mongodb) host")
		viper.BindPFlag("store_host", pflag.Lookup("store-host"))

		pflag.String("store-db", "argo_msg", "datastore (mongodb) database name")
		viper.BindPFlag("store_db", pflag.Lookup("store-db"))

		pflag.String("certificate", "/etc/pki/tls/certs/localhost.crt", "certificate file *.crt")
		viper.BindPFlag("certificate", pflag.Lookup("certificate"))

		pflag.String("certificate-key", "/etc/pki/tls/private/localhost.key", "certificate key file *.key")
		viper.BindPFlag("certificate_key", pflag.Lookup("certificate-key"))

		pflag.Bool("per-resource-auth", true, "enable per resource authentication")
		viper.BindPFlag("per_resource_auth", pflag.Lookup("per-resource-auth"))

		pflag.String("service-key", "", "service token definition for immediate full api access")
		viper.BindPFlag("service_key", pflag.Lookup("service-key"))

		pflag.String("push-enabled", "", "enable automatic handling of push subscriptions at start-up")
		viper.BindPFlag("push_enabled", pflag.Lookup("push-enabled"))

		configPath = pflag.String("config-dir", "", "directory path to an alternative json config file")

		pflag.Parse()

	}

	viper.SetConfigName("config")
	if configPath != nil {
		viper.AddConfigPath(*configPath)
	}
	viper.AddConfigPath("/etc/argo-messaging")
	viper.AddConfigPath(".")

	// Find and read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Errod trying to read the configuration file: %s \n", err))
	}

	// First check log level parameter and set logger
	cfg.LogLevel = viper.GetString("log_level")
	setLogLevel(cfg.LogLevel)
	log.Info("CONFIG", "\t", "Parameter Loaded - log_level: ", cfg.LogLevel)
	// Then load rest of the parameters
	cfg.BindIP = viper.GetString("bind_ip")
	log.Info("CONFIG", "\t", "Parameter Loaded - bind_ip: ", cfg.BindIP)
	cfg.Port = viper.GetInt("port")
	log.Info("CONFIG", "\t", "Parameter Loaded - port: ", cfg.Port)
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.Info("CONFIG", "\t", "Parameter Loaded - zookeeper_hosts: ", cfg.ZooHosts)
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.Info("CONFIG", "\t", "Parameter Loaded - kafka_znode: ", cfg.KafkaZnode)
	cfg.StoreHost = viper.GetString("store_host")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_host: ", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_db: ", cfg.StoreDB)
	cfg.Cert = viper.GetString("certificate")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate: ", cfg.Cert)
	cfg.CertKey = viper.GetString("certificate_key")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate_key: ", cfg.CertKey)
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.Info("CONFIG", "\t", "Parameter Loaded - per_resource_auth: ", cfg.ResAuth)
	cfg.ServiceToken = viper.GetString("service_token")
	log.Info("CONFIG", "\t", "Parameter Loaded - service_token: ", cfg.ServiceToken)
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.Info("CONFIG", "\t", "Parameter Loaded - push_enabled: ", cfg.PushEnabled)

}

// LoadStrJSON Loads configuration from a JSON string
func (cfg *APICfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(bytes.NewBuffer([]byte(input)))
	// Load Kafka configuration
	cfg.BindIP = viper.GetString("bind_ip")
	log.Info("CONFIG", "\t", "Parameter Loaded - bind_ip", cfg.BindIP)
	cfg.Port = viper.GetInt("port")
	log.Info("CONFIG", "\t", "Parameter Loaded - port", cfg.Port)
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.Info("CONFIG", "\t", "Parameter Loaded - zookeeper_hosts", cfg.ZooHosts)
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_host", cfg.KafkaZnode)
	cfg.StoreHost = viper.GetString("store_host")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_host", cfg.StoreHost)
	cfg.StoreDB = viper.GetString("store_db")
	log.Info("CONFIG", "\t", "Parameter Loaded - store_db", cfg.StoreDB)
	cfg.Cert = viper.GetString("certificate")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate", cfg.Cert)
	cfg.CertKey = viper.GetString("certificate_key")
	log.Info("CONFIG", "\t", "Parameter Loaded - certificate_key", cfg.CertKey)
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.Info("CONFIG", "\t", "Parameter Loaded - per_resource_auth", cfg.CertKey)
	cfg.ServiceToken = viper.GetString("service_token")
	log.Info("CONFIG", "\t", "Parameter Loaded - service_token", cfg.ServiceToken)
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.Info("CONFIG", "\t", "Parameter Loaded - push_enabled: ", cfg.PushEnabled)

}
