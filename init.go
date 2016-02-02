package main

import (
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
)

// Globals
var kafkaCfg = config.NewKafkaCfg()
var broker brokers.KafkaBroker

func init() {
	kafkaCfg.Load()
}
