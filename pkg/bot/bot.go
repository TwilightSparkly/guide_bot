package bot

// Ivan Orshak, 12.07.2023

import (
	"context"
	"encoding/json"
	tgWrapper "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"pocket_guide/pkg/broker"
	"time"
)

// NewBot is a method of the Bot structure receives a telegram api token from the environment variables,
// establishes a connection with the telegram server
// and fills in the fields of a variable of the Bot structure type
func (b *Bot) NewBot() error {
	// Logging layer
	b.log.NewLog("logs/bot/")

	// Broker layer
	b.err = b.newMsgBrk()
	if b.err != nil {
		b.log.LogErr.Println("NewBot(): Unable to create broker, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBot(): Broker has been successfully created.")
	}

	// Trying to get telegram bot token
	apiKey, flag := os.LookupEnv("BOT_TOKEN")
	if !flag {
		b.log.LogErr.Println("NewBot(): BOT_TOKEN env variable not found.")
		return b.err
	} else {
		b.log.LogInfo.Println("NewBot(): The token was successfully received.")
	}

	// Making the connection to telegram server
	b.bot, b.err = tgWrapper.NewBotAPI(apiKey)
	if b.err != nil {
		b.log.LogErr.Println("NewBot(): Unable to authorize on account, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBot(): Authorized on account:", b.bot.Self.UserName)
	}

	return nil
}

// Close closes the logging system and shuts down the broker
func (b *Bot) Close() {
	defer b.log.Close()
	defer b.Consumer.Close()
	defer b.Producer.Close()
}

// newMsgBrk creates a consumer/producer pair
// and two queues required to work with the broker
func (b *Bot) newMsgBrk() error {
	//Consumer initialization
	b.err = b.Consumer.NewBroker()
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to create consumer, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): Consumer has been successfully created.")
	}

	// Making consumer queue
	b.err = b.Consumer.MakeQueue("Response")
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to create a consumer queue, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): A consumer queue has been successfully created.")
	}

	// Producer initialization
	b.err = b.Producer.NewBroker()
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to create producer, error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): Producer has been successfully created.")
	}

	// Creating producer queues to sending requests
	b.err = b.Producer.MakeQueue("aiRequest")
	if b.err != nil {
		b.log.LogErr.Println("NewBroker(): Unable to create a queue 'aiRequest', error:", b.err)
		return b.err
	} else {
		b.log.LogInfo.Println("NewBroker(): A producer queue 'aiRequest' has been successfully created.")
	}

	return nil
}

// Listener is a method of the Bot structure endlessly listens to the message channel
// from the telegram server and sends each update to the handler
func (b *Bot) Listener() {
	u := tgWrapper.NewUpdate(0)
	u.Timeout = 60

	// Making listening channel
	updates := b.bot.GetUpdatesChan(u)

	// For each update from telegram run handler in goroutine
	for update := range updates {
		go func(update tgWrapper.Update) {
			err := b.handleMsg(update)
			if err != nil {
				u, _ := json.Marshal(update)
				b.log.LogErr.Println("Listener(): Unable to handle update: ", u, " error: ", err)
			}
		}(update)
	}
}

// Sender is a method of the Bot structure listens to the broker's channel
// and sends incoming messages to the telegram server
func (b *Bot) Sender() error {
	// Channel to connect Sender() with consumer
	ch := make(chan broker.UserMsg)

	// Start listening to the broker's channel
	go func() {
		err := b.Consumer.Consume("Response", ch)
		if err != nil {
			b.log.LogErr.Println("Sender(): Cannot consume messages, error:", err)
		}
	}()

	// If a message is received from the broker
	for {
		data, _ := <-ch
		// A goroutine is created for each incoming message
		go func(data broker.UserMsg) {
			if len(data.Data) != 0 {
				// Creating a variable with the desired type to send to the telegram server via API
				msg := tgWrapper.NewMessage(data.ChatId.Id, data.Data)

				// Sending a message
				_, err := b.bot.Send(msg)
				if err != nil {
					b.log.LogErr.Println("Sender(): Unable to send a message to telegram, error:", err)
				}
			}
		}(data)
	}
}

// handleMsg is a method that contains business logic
// and allows you to separate command messages from ordinary ones.
// Ordinary messages are sent by the broker to the microservice for working with AI,
// command messages are sent to their own handler
func (b *Bot) handleMsg(update tgWrapper.Update) error {
	// Background context to broker
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// If we got a message
	if update.Message != nil {
		// If it is a command message: '/command'
		if update.Message.IsCommand() {

		} else { // If we got a standard message - send to AI service
			// Creating a variable with the desired type to send to the telegram server via API
			msg := tgWrapper.NewMessage(update.Message.Chat.ID, "Дайте подумать...")
			var data []byte

			data, err := json.Marshal(update.Message)
			if err != nil {
				b.log.LogErr.Println("handleMsg(): Unable to convert into json, error:", err)
				return err
			}

			// Sending a notification about request processing
			_, err = b.bot.Send(msg)
			if err != nil {
				b.log.LogErr.Println("handleMsg(): Unable to send a message to telegram, error:", err)
			}

			// Trying to publish message to AI service
			err = b.Producer.Publish(data, "aiRequest", ctx)
			if err != nil {
				b.log.LogErr.Println("handleMsg(): Unable to publish message to AI service, error:", err)

				// Creating a variable with the desired type to send to the telegram server via API
				msg = tgWrapper.NewMessage(update.Message.Chat.ID, "Извините, сервис для общения с искусственным "+
					"интеллектом временно не работает.")
				_, err = b.bot.Send(msg)
				if err != nil {
					b.log.LogErr.Println("handleMsg(): Unable to send a message to telegram, error:", err)
				}
				return err
			}
		}
	}

	return nil
}
