package main

import (
	"log"
	"net/http"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
)

func main() {
	// create and load configuration object
	cfg := config.NewAPICfg("LOAD")

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.BrokerHost)
	defer broker.CloseConnections()

	// create the store
	store := stores.NewMongoStore(cfg.StoreHost, cfg.StoreDB)

	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, defaultRoutes)

	// Start http server listening using API.Router
	log.Fatal(http.ListenAndServe(":8080", API.Router))
}
