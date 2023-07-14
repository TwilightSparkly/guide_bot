package bot

// Ivan Orshak, 12.07.2023

import (
	tgWrapper "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pocket_guide/pkg/broker"
	"pocket_guide/pkg/logging"
)

type Bot struct {
	bot      *tgWrapper.BotAPI
	Consumer broker.Broker
	Producer broker.Broker
	log      logging.Log
	err      error
}
