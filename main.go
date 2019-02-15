package main

import (
	"crypto/tls"
	"log/syslog"

	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	push2 "github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

// setup logrus' std logger with syslog hook
func init() {
	// dont use colors in output
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		log.AddHook(hook)
	}
}

func main() {
	// create and load configuration object
	cfg := config.NewAPICfg("LOAD")

	// create the store
	store := stores.NewMongoStore(cfg.StoreHost, cfg.StoreDB)
	store.Initialize()

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.GetBrokerInfo())
	defer broker.CloseConnections()

	mgr := &push2.Manager{}

	// ams push server pushClient
	pushClient := push.NewGrpcClient(cfg)

	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, mgr, pushClient, defaultRoutes)

	//Configure TLS support only
	config := &tls.Config{
		MinVersion:               tls.VersionTLS10,
		PreferServerCipherSuites: true,
	}

	// Initialize CORS specifics
	xReqWithConType := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	allowVerbs := handlers.AllowedMethods([]string{"OPTIONS", "POST", "GET", "PUT", "DELETE", "HEAD"})
	// Initialize server wth proper parameters
	server := &http.Server{Addr: ":" + strconv.Itoa(cfg.Port), Handler: handlers.CORS(xReqWithConType, allowVerbs)(API.Router), TLSConfig: config}

	// Web service binds to server. Requests served over HTTPS.
	err := server.ListenAndServeTLS(cfg.Cert, cfg.CertKey)
	if err != nil {
		log.Fatal("API", "\t", "ListenAndServe:", err)
	}

}
