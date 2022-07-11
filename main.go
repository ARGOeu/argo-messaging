package main

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/version"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func init() {
	// don't use colors in output
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp: true,
			DisableColors: true},
	)

	// display binary version information during start up
	version.LogInfo()

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

	mgr := &oldPush.Manager{}

	// ams push server pushClient
	pushClient := push.NewGrpcClient(cfg)
	err := pushClient.Dial()
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "ams-push-server",
				"backend_hosts":   pushClient.Target(),
			},
		).Error(err.Error())
	}

	defer pushClient.Close()

	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, mgr, pushClient, defaultRoutes)

	//Configure TLS support only
	config := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	// Initialize CORS specifics
	xReqWithConType := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "x-api-key"})
	allowVerbs := handlers.AllowedMethods([]string{"OPTIONS", "POST", "GET", "PUT", "DELETE", "HEAD"})
	// Initialize server wth proper parameters
	server := &http.Server{Addr: ":" + strconv.Itoa(cfg.Port), Handler: handlers.CORS(xReqWithConType, allowVerbs)(API.Router), TLSConfig: config}

	// Web service binds to server. Requests served over HTTPS.
	err = server.ListenAndServeTLS(cfg.Cert, cfg.CertKey)
	if err != nil {
		log.Fatal("API", "\t", "ListenAndServe:", err)
	}

}
