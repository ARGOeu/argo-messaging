package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"crypto/x509"
	"github.com/samuel/go-zookeeper/zk"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"
)

// AuthOption defines how the service will handle authentication/authorization
// KEY, HEADER or BOTH are the available values for where the auth token should reside
type AuthOption int

const (
	// the api key can reside in the url parameter 'key'
	// maps to config value 'key'
	UrlKey = iota + 1
	// the api key can reside in the header 'x-api-key'
	// maps to config value 'header'
	HeaderKey
	// the api key can reside in either of the two
	// maps to config value 'both'
	URLKeyAndHeaderKey
)

// String representation of the iota auth option
func (a AuthOption) String() string {
	return [...]string{"key", "header", "both"}[a-1]
}

// APICfg holds kafka configuration
type APICfg struct {
	// values
	BindIP                    string
	Port                      int
	ZooHosts                  []string
	KafkaZnode                string //The Zookeeper znode used by Kafka
	StoreHost                 string
	StoreDB                   string
	Cert                      string
	CertKey                   string
	CertificateAuthoritiesDir string
	ResAuth                   bool
	ServiceToken              string
	LogLevel                  string
	PushEnabled               bool
	// Whether or not it should communicate over tls with the push server
	PushTlsEnabled bool
	// Push server endpoint
	PushServerHost string
	// Push server port
	PushServerPort int
	// If tls is enabled, whether or not it should verify the push server's certificate
	VerifyPushServer bool
	// The token that corresponds to the registered push worker user
	PushWorkerToken string
	// Logging output(console,file,syslog etc)
	LogFacilities []string
	// AuthOption defines how the service will handle authentication/authorization
	// KEY, HEADER or BOTH are the available values for where the auth token should reside
	authOption AuthOption
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
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "zookeeper",
					"error":           err.Error(),
				},
			).Error("Invalid broker list")
		} else {
			// Info retrieved succesfully so continue
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "zookeeper",
					"broker_list":     brkList,
				},
			).Info("Broker list discovered")

			return brkList
		}

	}
}

// GetZooList gets broker list from zookeeper
func (cfg *APICfg) GetZooList() ([]string, error) {

	peerList := []string{}

	log.WithFields(
		log.Fields{
			"type":            "backend_log",
			"backend_service": "zookeeper",
			"backend_hosts":   cfg.ZooHosts,
		},
	).Info("Trying to connect to Zookeeper")

	zConn, _, err := zk.Connect(cfg.ZooHosts, time.Second)
	// Check if indeed connected and can read
	_, _, _, err = zConn.ChildrenW("/")
	if err != nil {
		zConn.Close()
		return peerList, err
	}

	log.WithFields(
		log.Fields{
			"type":            "backend_log",
			"backend_service": "zookeeper",
			"backend_hosts":   cfg.ZooHosts,
		},
	).Info("Connection to Zookeeper established successfully")

	log.WithFields(
		log.Fields{
			"type":            "backend_log",
			"backend_service": "zookeeper",
			"backend_hosts":   cfg.ZooHosts,
		},
	).Info("Attempting to read broker information")

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

// LoadCAs builds the CA chain using pem files from the specified directory in the cfg
func (cfg *APICfg) LoadCAs() (roots *x509.CertPool) {

	log.Info("Building the root CA chain...")

	pattern := "*.pem"
	roots = x509.NewCertPool()

	err := filepath.Walk(cfg.CertificateAuthoritiesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Prevent panic by handling failure accessing a path %q: %v\n", cfg.CertificateAuthoritiesDir, err)
			return err
		}

		if ok, _ := filepath.Match(pattern, info.Name()); ok {
			bytes, err := ioutil.ReadFile(filepath.Join(cfg.CertificateAuthoritiesDir, info.Name()))
			if err != nil {
				return err
			}

			if ok = roots.AppendCertsFromPEM(bytes); !ok {
				return fmt.Errorf("Could not append cert to CA: %v ", filepath.Join(cfg.CertificateAuthoritiesDir, info.Name()))
			}
		}

		return nil
	})

	if err != nil {
		log.Errorf("error walking the path %q: %v\n", cfg.CertificateAuthoritiesDir, err)
	}

	log.Info("All certificates parsed successfully.")

	return roots
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

func setLogFacilities(facilities []string) {

	if len(facilities) == 0 {
		return
	}

	consoleEnabled := false

	for _, f := range facilities {

		if strings.ToUpper(f) == "SYSLOG" {
			hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
			if err == nil {
				log.AddHook(hook)
			} else {
				log.Errorf("Couldn't set up syslog handler, %v", err.Error())
			}
		}

		if strings.ToUpper(f) == "CONSOLE" {
			consoleEnabled = true
		}
	}

	// if the console option has not been specified close the standard logging
	if !consoleEnabled {
		log.SetOutput(ioutil.Discard)
	}
}

// setAuthOption determines which auth option should be used
func (cfg *APICfg) setAuthOption(authOpt string) {

	switch strings.ToLower(authOpt) {
	case "both":
		cfg.authOption = URLKeyAndHeaderKey
		break
	case "header":
		cfg.authOption = HeaderKey
		break
	default:
		cfg.authOption = UrlKey
	}
}

// AuthOption returns the value of the config for auth_option
func (cfg *APICfg) AuthOption() AuthOption {
	return cfg.authOption
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
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - log_level: %v", cfg.LogLevel)

	cfg.LogFacilities = viper.GetStringSlice("log_facilities")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - log_facilities: %v", cfg.LogFacilities)
	setLogFacilities(cfg.LogFacilities)

	// Then load rest of the parameters
	cfg.setAuthOption(viper.GetString("auth_option"))
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - auth_option: %v", cfg.AuthOption())

	// bind ip
	cfg.BindIP = viper.GetString("bind_ip")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - bind_ip: %v", cfg.BindIP)

	// service port
	cfg.Port = viper.GetInt("port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - port: %v", cfg.Port)

	// zookeeper hosts
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - zookeeper_hosts: %v", cfg.ZooHosts)

	// kafka_znode
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - kafka_znode: %v", cfg.KafkaZnode)

	// store host
	cfg.StoreHost = viper.GetString("store_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_host: %v", cfg.StoreHost)

	// store name
	cfg.StoreDB = viper.GetString("store_db")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_db: %v", cfg.StoreDB)

	// service certificate
	cfg.Cert = viper.GetString("certificate")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate: %v", cfg.Cert)

	// service certificate key
	cfg.CertKey = viper.GetString("certificate_key")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_key: %v", cfg.CertKey)

	// certificate authorities directory
	cfg.CertificateAuthoritiesDir = viper.GetString("certificate_authorities_dir")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_authorities_dir: %v", cfg.CertificateAuthoritiesDir)

	// per resource authorisation
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - per_resource_auth: %v", cfg.ResAuth)

	// service token
	cfg.ServiceToken = viper.GetString("service_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - service_token")

	// push enabled true or false
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_enabled: %v", cfg.PushEnabled)

	// push TLS enabled true or false
	cfg.PushTlsEnabled = viper.GetBool("push_tls_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_tls_enabled: %v", cfg.PushTlsEnabled)

	// push server host
	cfg.PushServerHost = viper.GetString("push_server_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_host: %v", cfg.PushServerHost)

	// push server port
	cfg.PushServerPort = viper.GetInt("push_server_port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_port: %v", cfg.PushServerPort)

	// verify push server
	cfg.VerifyPushServer = viper.GetBool("verify_push_server")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - verify_push_server: %v", cfg.VerifyPushServer)

	// push worker token
	cfg.PushWorkerToken = viper.GetString("push_worker_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - push_worker_token")
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

		pflag.String("ca-dir", "/etc/grid-security/certificates", "directory containing the ca files *.pem")
		viper.BindPFlag("certificate_authorities_dir", pflag.Lookup("ca-dir"))

		pflag.Bool("per-resource-auth", true, "enable per resource authentication")
		viper.BindPFlag("per_resource_auth", pflag.Lookup("per-resource-auth"))

		pflag.String("service-key", "", "service token definition for immediate full api access")
		viper.BindPFlag("service_key", pflag.Lookup("service-key"))

		pflag.String("push-enabled", "", "enable automatic handling of push subscriptions at start-up")
		viper.BindPFlag("push_enabled", pflag.Lookup("push-enabled"))

		pflag.Bool("push-tls", true, "enable tls for communicating withe ams push server")
		viper.BindPFlag("push_tls_enabled", pflag.Lookup("push-tls"))

		pflag.String("push-host", "", "push server hostname")
		viper.BindPFlag("push_server_host", pflag.Lookup("push-host"))

		pflag.Int("push-port", 0, "push server port")
		viper.BindPFlag("push_server_port", pflag.Lookup("push-port"))

		pflag.Bool("push-verify", true, "verify push server's certificate if tls is enabled")
		viper.BindPFlag("verify_push_server", pflag.Lookup("push-verify"))

		pflag.String("push-worker-token", "", "token corresponding to the registered push worker user")
		viper.BindPFlag("push_worker_token", pflag.Lookup("push-worker-token"))

		pflag.String("log-facilities", "", "logging output(s)")
		viper.BindPFlag("log_facilities", pflag.Lookup("log-facilities"))

		pflag.String("auth-option", "", "where the auth token should reside")
		viper.BindPFlag("auth_option", pflag.Lookup("auth-option"))

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
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - log_level: %v", cfg.LogLevel)

	cfg.LogFacilities = viper.GetStringSlice("log_facilities")
	setLogFacilities(cfg.LogFacilities)
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - log_facilities: %v", cfg.LogFacilities)

	// Then load rest of the parameters

	cfg.setAuthOption(viper.GetString("auth_option"))
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - auth_option: %v", cfg.AuthOption())

	// bind ip
	cfg.BindIP = viper.GetString("bind_ip")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - bind_ip: %v", cfg.BindIP)

	// service port
	cfg.Port = viper.GetInt("port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - port: %v", cfg.Port)

	// zookeeper hosts
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - zookeeper_hosts: %v", cfg.ZooHosts)

	// kafka_znode
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - kafka_znode: %v", cfg.KafkaZnode)

	// store host
	cfg.StoreHost = viper.GetString("store_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_host: %v", cfg.StoreHost)

	// store name
	cfg.StoreDB = viper.GetString("store_db")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_db: %v", cfg.StoreDB)

	// service certificate
	cfg.Cert = viper.GetString("certificate")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate: %v", cfg.Cert)

	// service certificate key
	cfg.CertKey = viper.GetString("certificate_key")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_key: %v", cfg.CertKey)

	// certificate authorities directory
	cfg.CertificateAuthoritiesDir = viper.GetString("certificate_authorities_dir")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_authorities_dir: %v", cfg.CertificateAuthoritiesDir)

	// per resource authorisation
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - per_resource_auth: %v", cfg.ResAuth)

	// service token
	cfg.ServiceToken = viper.GetString("service_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - service_token")

	// push enabled true or false
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_enabled: %v", cfg.PushEnabled)

	// push TLS enabled true or false
	cfg.PushTlsEnabled = viper.GetBool("push_tls_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_tls_enabled: %v", cfg.PushTlsEnabled)

	// push server host
	cfg.PushServerHost = viper.GetString("push_server_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_host: %v", cfg.PushServerHost)

	// push server port
	cfg.PushServerPort = viper.GetInt("push_server_port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_port: %v", cfg.PushServerPort)

	// verify push server
	cfg.VerifyPushServer = viper.GetBool("verify_push_server")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - verify_push_server: %v", cfg.VerifyPushServer)

	// push worker token
	cfg.PushWorkerToken = viper.GetString("push_worker_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - push_worker_token")

}

// LoadStrJSON Loads configuration from a JSON string
func (cfg *APICfg) LoadStrJSON(input string) {
	viper.SetConfigType("json")
	viper.ReadConfig(strings.NewReader(input))
	// Load Kafka configuration
	// bind ip
	cfg.BindIP = viper.GetString("bind_ip")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - bind_ip: %v", cfg.BindIP)

	// service port
	cfg.Port = viper.GetInt("port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - port: %v", cfg.Port)

	// zookeeper hosts
	cfg.ZooHosts = viper.GetStringSlice("zookeeper_hosts")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - zookeeper_hosts: %v", cfg.ZooHosts)

	// kafka_znode
	cfg.KafkaZnode = viper.GetString("kafka_znode")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - kafka_znode: %v", cfg.KafkaZnode)

	// store host
	cfg.StoreHost = viper.GetString("store_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_host: %v", cfg.StoreHost)

	// store name
	cfg.StoreDB = viper.GetString("store_db")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - store_db: %v", cfg.StoreDB)

	// service certificate
	cfg.Cert = viper.GetString("certificate")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate: %v", cfg.Cert)

	// service certificate key
	cfg.CertKey = viper.GetString("certificate_key")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_key: %v", cfg.CertKey)

	// certificate authorities directory
	cfg.CertificateAuthoritiesDir = viper.GetString("certificate_authorities_dir")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - certificate_authorities_dir: %v", cfg.CertificateAuthoritiesDir)

	// per resource authorisation
	cfg.ResAuth = viper.GetBool("per_resource_auth")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - per_resource_auth: %v", cfg.ResAuth)

	// service token
	cfg.ServiceToken = viper.GetString("service_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - service_token:")

	// push enabled true or false
	cfg.PushEnabled = viper.GetBool("push_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_enabled: %v", cfg.PushEnabled)

	// push TLS enabled true or false
	cfg.PushTlsEnabled = viper.GetBool("push_tls_enabled")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_tls_enabled: %v", cfg.PushTlsEnabled)

	// push server host
	cfg.PushServerHost = viper.GetString("push_server_host")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_host: %v", cfg.PushServerHost)

	// push server port
	cfg.PushServerPort = viper.GetInt("push_server_port")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - push_server_port: %v", cfg.PushServerPort)

	// verify push server
	cfg.VerifyPushServer = viper.GetBool("verify_push_server")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - verify_push_server: %v", cfg.VerifyPushServer)

	// push worker token
	cfg.PushWorkerToken = viper.GetString("push_worker_token")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("Parameter Loaded - push_worker_token")

	cfg.LogFacilities = viper.GetStringSlice("log_facilities")
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - log_facilities: %v", cfg.LogFacilities)

	// auth option
	cfg.setAuthOption(viper.GetString("auth_option"))
	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Infof("Parameter Loaded - auth_option: %v", cfg.AuthOption())
}
