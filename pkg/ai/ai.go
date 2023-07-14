package ai

// Ivan Orshak, 13.07.2023

import (
	"github.com/otiai10/openaigo"
	"os"
)

// NewAi Ai method connects the logging system to the object,
// initializes the API key from the environment variables,
// connects to the artificial intelligence server using the API key,
// and also creates one producer and one consumer to work with the broker
func (a *Ai) NewAi() error {
	// Logging layer
	a.log.NewLog("logs/ai/")

	//Trying to get api key from env vars
	apiKey, flag := os.LookupEnv("GPT_TOKEN")
	if !flag {
		a.log.LogErr.Println("NewAi(): BOT_TOKEN env variable not found.")
		return a.err
	} else {
		a.log.LogInfo.Println("NewAi(): The token was successfully received.")
	}

	// Trying to make connection to AI servers
	a.Client = openaigo.NewClient(apiKey)

	// Creating broker objects
	a.err = a.newMsgBrk()
	if a.err != nil {
		a.log.LogErr.Println("NewAi(): Unable to create broker, error:", a.err)
		return a.err
	} else {
		a.log.LogInfo.Println("NewAi(): Broker has been successfully created.")
	}

	return nil
}

// Close shuts down the logging system and disconnects from the broker
func (a *Ai) Close() {
	defer a.log.Close()
	defer a.Consumer.Close()
	defer a.Producer.Close()
}

// MakeRequest fills in the fields of the structure type variable
// required to send the API request
func (a *Ai) MakeRequest(data string) openaigo.ChatRequest {
	// Убрать в конфиги
	request := openaigo.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openaigo.Message{
			{Role: "user", Content: data},
		},
	}

	return request
}

// newMsgBrk creates a consumer/producer pair
// and two queues required to work with the broker
func (a *Ai) newMsgBrk() error {
	a.err = a.Consumer.NewBroker()
	if a.err != nil {
		a.log.LogErr.Println("newMsgBrk(): Unable to create consumer, error:", a.err)
		return a.err
	} else {
		a.log.LogInfo.Println("newMsgBrk(): Consumer has been successfully created.")
	}

	a.err = a.Producer.NewBroker()
	if a.err != nil {
		a.log.LogErr.Println("newMsgBrk(): Unable to create producer, error:", a.err)
		return a.err
	} else {
		a.log.LogInfo.Println("newMsgBrk(): Producer has been successfully created.")
	}

	a.err = a.Consumer.MakeQueue("aiRequest")
	if a.err != nil {
		a.log.LogErr.Println("newMsgBrk(): Unable to create a queue 'aiRequest', error:", a.err)
		return a.err
	} else {
		a.log.LogInfo.Println("newMsgBrk(): A consumer queue 'aiRequest' has been successfully created.")
	}

	a.err = a.Producer.MakeQueue("Response")
	if a.err != nil {
		a.log.LogErr.Println("newMsgBrk(): Unable to create a queue 'Response', error:", a.err)
		return a.err
	} else {
		a.log.LogInfo.Println("newMsgBrk(): A producer queue 'Response' has been successfully created.")
	}

	return nil
}
