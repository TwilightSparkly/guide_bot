package ai

// Ivan Orshak, 13.07.2023

import (
	"github.com/otiai10/openaigo"
	"pocket_guide/pkg/broker"
	"pocket_guide/pkg/logging"
)

type Ai struct {
	Client   *openaigo.Client
	apiKey   string
	Consumer broker.Broker
	Producer broker.Broker
	log      logging.Log
	err      error
}
