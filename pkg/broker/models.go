package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"pocket_guide/pkg/logging"
)

// Ivan Orshak, 13.07.2023

type Broker struct {
	ch     *amqp.Channel
	conn   *amqp.Connection
	queues map[string]amqp.Queue
	log    logging.Log
	cfgF   *os.File
	err    error
}

type UserMsg struct {
	Data   string `json:"text"`
	ChatId struct {
		Id int64 `json:"id"`
	} `json:"from"`
}
