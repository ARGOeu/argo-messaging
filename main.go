package main

import (
	"log"
	"net/http"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
)

func main() {
	// create and load configuration object
	cfg := config.NewKafkaCfg("LOAD")

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.Server)
	defer broker.CloseConnections()

	// create and initialize API routing object
	API := NewRouting(cfg, broker, defaultRoutes)

	// Start http server listening using API.Router
	log.Fatal(http.ListenAndServe(":8080", API.Router))
}
