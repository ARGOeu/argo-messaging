package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/push"
	"github.com/ARGOeu/argo-messaging/stores"
)

func main() {
	// create and load configuration object
	cfg := config.NewAPICfg("LOAD")

	// create the store
	store := stores.NewMongoStore(cfg.StoreHost, cfg.StoreDB)
	store.Initialize()

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.BrokerHosts)
	defer broker.CloseConnections()

	sndr := push.NewHTTPSender(1)

	mgr := push.NewManager(broker, store, sndr)
	mgr.LoadPushSubs()
	mgr.StartAll()
	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, mgr, defaultRoutes)

	//Configure TLS support only
	config := &tls.Config{
		MinVersion: tls.VersionTLS10,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		},
		PreferServerCipherSuites: true,
	}

	// Initialize server wth proper parameters
	server := &http.Server{Addr: ":" + strconv.Itoa(cfg.Port), Handler: API.Router, TLSConfig: config}

	// Web service binds to server. Requests served over HTTPS.
	err := server.ListenAndServeTLS(cfg.Cert, cfg.CertKey)
	if err != nil {
		log.Fatal("ERROR\tAPI\tListenAndServe:", err)
	}

}
