package broker

// Ivan Orshak, 12.07.2023

import (
	"bufio"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"strconv"
	"strings"
)

// NewBroker is a method that initializes its own logging system,
// creates a connection and a channel with the broker,
// loads configuration file
func (b *Broker) NewBroker() error {
	// Logging layer
	b.log.NewLog("logs/broker/")

	// Initializing map queues
	b.queues = make(map[string]amqp.Queue)

	// Trying to connect over TCP with broker service
	var brokerUrl string
	brokerUrl, b.err = b.getConfig("BROKER_URL")

	b.conn, b.err = amqp.Dial(brokerUrl)
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to connect over TCP, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): Connected to:", brokerUrl)
	}

	// Trying to create a connection channel with broker service
	b.ch, b.err = b.conn.Channel()
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to create a connection channel, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): Connection successfully created.")
	}

	return nil
}

// MakeQueue is a method that creates a queue in the queue map
// for writing and reading broker messages with the name passed in the method parameter
func (b *Broker) MakeQueue(qname string) error {
	b.queues[qname], b.err = b.ch.QueueDeclare(
		qname, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if b.err != nil {
		b.log.LogErr.Println("MakeQueue(): Unable to create a queue:", qname, "error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("MakeQueue(): A queue:", qname, "has been successfully created.")
	}

	return nil
}

// Publish method sends a message to the broker in the queue specified in the input parameters
func (b *Broker) Publish(msg []byte, qname string, ctx context.Context) error {
	b.err = b.ch.PublishWithContext(ctx,
		"",                   // exchange
		b.queues[qname].Name, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	if b.err != nil {
		b.log.LogErr.Println("Publish(): Unable to publish a message, error:", b.err)
		return b.err
	}

	return nil
}

// Consume method infinitely checks for incoming messages
// from the broker in the queue with the name passed in the input parameters
func (b *Broker) Consume(qname string, ch chan UserMsg) error {
	// Channel to connect Consume() with internal goroutine
	var forever chan struct{}
	// Channel for incoming broker messages
	var messages <-chan amqp.Delivery
	messages, b.err = b.makeConsumeCh(qname)
	if b.err != nil {
		b.log.LogErr.Println("Consume(): Unable to consume publishing messages, error:", b.err)
		return b.err
	}

	// Goroutine for processing each incoming message
	go func(ch chan UserMsg) {
		for message := range messages {
			var msg UserMsg

			b.err = json.Unmarshal(message.Body, &msg)
			if b.err != nil {
				b.log.LogErr.Println("Consume(): Unable to convert from json, error:", b.err)
			} else {
				ch <- msg
			}
		}
	}(ch)

	<-forever

	return nil
}

// makeConsumeCh is a method that creates a connection channel
// to the broker to listen to incoming messages
func (b *Broker) makeConsumeCh(qname string) (<-chan amqp.Delivery, error) {
	var params [6]string
	var autoAck, exclusive, noLocal, noWait bool
	var args amqp.Table

	params[0], b.err = b.getConfig("BROKER_URL")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER conf, error:", b.err)
		return nil, b.err
	}
	params[1], b.err = b.getConfig("BROKER_CONSUMER_AUTO_ACK")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER_AUTO_ACK conf, error:", b.err)
		return nil, b.err
	}
	params[2], b.err = b.getConfig("BROKER_CONSUMER_EXCLUSIVE")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER_EXCLUSIVE conf, error:", b.err)
		return nil, b.err
	}
	params[3], b.err = b.getConfig("BROKER_CONSUMER_NO_LOCAL")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER_NO_LOCAL conf, error:", b.err)
		return nil, b.err
	}
	params[4], b.err = b.getConfig("BROKER_CONSUMER_NO_WAIT")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER_NO_WAIT conf, error:", b.err)
		return nil, b.err
	}
	params[5], b.err = b.getConfig("BROKER_CONSUMER_ARGS")
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to get BROKER_CONSUMER_ARGS conf, error:", b.err)
		return nil, b.err
	}

	autoAck, b.err = strconv.ParseBool(params[1])
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to parse bool parameter:", params[4], "error:", b.err)
		return nil, b.err
	}
	exclusive, b.err = strconv.ParseBool(params[2])
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to parse bool parameter", params[4], "error:", b.err)
		return nil, b.err
	}
	noLocal, b.err = strconv.ParseBool(params[3])
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to parse bool parameter", params[4], "error:", b.err)
		return nil, b.err
	}
	noWait, b.err = strconv.ParseBool(params[4])
	if b.err != nil {
		b.log.LogErr.Println("makeConsumeCh(): Unable to parse bool parameter", params[4], "error:", b.err)
		return nil, b.err
	}

	return b.ch.Consume(
		b.queues[qname].Name, // queue
		params[0],            // consumer
		autoAck,              // auto-ack
		exclusive,            // exclusive
		noLocal,              // no-local
		noWait,               // no-wait
		args,                 // args
	)
}

// Close method closes the connection with the broker and the connection channel
// and closes configuration file
func (b *Broker) Close() {
	// Trying to close broker connection
	b.err = b.conn.Close()
	if b.err != nil {
		b.log.LogErr.Println("Close(): Unable to close the connection, error:", b.err)
	} else {
		b.log.LogInfo.Println("Close(): The connection was successfully closed.")
	}

	// Trying to close broker channel
	b.err = b.ch.Close()
	if b.err != nil {
		b.log.LogErr.Println("Close(): Unable to close the channel, error:", b.err)
	} else {
		b.log.LogInfo.Println("Close(): The channel was successfully closed.")
	}
}

// loadConfig is a method that opens a configuration file
// and searches for the value of the parameter passed in the input parameters
func (b *Broker) getConfig(param string) (string, error) {
	// Trying to load configuration file
	b.cfgF, b.err = os.OpenFile("cfg/.cfg", os.O_RDONLY, 0666)
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to open configuration file, error:", b.err)
		return "", b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): Configuration file has been successfully loaded.")
	}
	defer func() {
		// Closing file after work
		b.err = b.cfgF.Close()
		if b.err != nil {
			b.log.LogErr.Println("Unable to close configuration file, error:", b.err)
		} else {
			b.log.LogInfo.Println("Close(): The conf file was successfully closed.")
		}
	}()

	// Creating the new scanner
	fileScanner := bufio.NewScanner(b.cfgF)
	// Read string by string
	for fileScanner.Scan() {
		data := strings.Split(fileScanner.Text(), "=")
		if data[0] == param {
			return data[1], nil
		}
	}

	// If there is not such parameter
	b.log.LogErr.Println("getConfig(): Parameter is not found.")
	//return "", errors.New("getConfig(): parameter is not found")
	return "", nil
}
