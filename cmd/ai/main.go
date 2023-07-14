package main

// Ivan Orshak, 13.07.2023

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	"pocket_guide/pkg/ai"
	"pocket_guide/pkg/broker"
	"pocket_guide/pkg/logging"
)

// Loading values from .env into the system
func init() {
	var log logging.Log
	log.NewLog("logs/ai/")

	err := godotenv.Load("cfg/.env")
	if err != nil {
		log.LogFatal.Fatal("init(): Cannot read env vars, bot have been stopped, error: ", err)
	} else {
		log.LogInfo.Println("init(): Env vars have been successfully loaded.")
	}
}

// The entry point to the program.
// Creating a broker, listening to the message channel,
// processing incoming messages in an infinite loop
func main() {
	// Logging layer
	var log logging.Log
	log.NewLog("logs/ai/")
	defer log.Close()
	var err error

	// Channel to connect main() with consumer goroutine
	ch := make(chan broker.UserMsg)

	// Creating broker
	var a ai.Ai
	err = a.NewAi()
	if err != nil {
		log.LogFatal.Fatal("main(): Unable to create ai object, error: ", err)
	}
	defer a.Close()

	//Listening to the broker's channel in goroutine
	go func() {
		err = a.Consumer.Consume("aiRequest", ch)
		if err != nil {
			log.LogFatal.Fatal("main(): Cannot consume messages, error:", err)
		}
	}()

	// If a message is received from the broker
	for {
		msg, _ := <-ch

		if len(msg.Data) != 0 {
			// Creating a request for AI
			request := a.MakeRequest(msg.Data)
			ctx := context.Background()

			//Parsing a request for AI, processing the response and publishing it in the Sender()
			go func(msg broker.UserMsg) {
				//Get response from AI API
				response, err := a.Client.Chat(ctx, request)
				if err != nil {
					msg.Data = "Извините, сервис для общения с искусственным интеллектом временно не работает."

					data, err := json.Marshal(msg)
					if err != nil {
						log.LogErr.Println("main(): Unable to convert into json, error:", err)
						return
					}

					err = a.Producer.Publish(data, "Response", ctx)
					if err != nil {
						log.LogErr.Println("main(): Unable to publish message to Sender(), error:", err)
						return
					}
				} else {
					msg.Data = response.Choices[0].Message.Content

					data, err := json.Marshal(msg)
					if err != nil {
						log.LogErr.Println("main(): Unable to convert into json, error:", err)
						return
					}

					err = a.Producer.Publish(data, "Response", ctx)
					if err != nil {
						log.LogErr.Println("main(): Unable to publish message to Sender(), error:", err)
						return
					}
				}
			}(msg)
		}
	}
}
