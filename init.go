package main

import (
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
)

// Globals
var broker brokers.KafkaBroker

func init() {
	config.Load()
}
