package main

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
)

func main() {
	// create and load configuration object
	cfg := config.NewAPICfg("LOAD")

	// create the store
	store := stores.NewMongoStore(cfg.StoreHost, cfg.StoreDB)
	store.Initialize()

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.GetBrokerInfo())
	defer broker.CloseConnections()

	sndr := push.NewHTTPSender(1)

	mgr := push.NewManager(broker, store, sndr)
	mgr.LoadPushSubs()
	mgr.StartAll()
	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, mgr, defaultRoutes)

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
