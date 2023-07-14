package main

// Ivan Orshak, 12.07.2023

import (
	"github.com/joho/godotenv"
	"pocket_guide/pkg/bot"
	"pocket_guide/pkg/logging"
)

// Loading values from .env into the system
func init() {
	var log logging.Log
	log.NewLog("logs/bot/")

	err := godotenv.Load("cfg/.env")
	if err != nil {
		log.LogFatal.Fatal("init(): Cannot read env vars, bot have been stopped, error: ", err)
	} else {
		log.LogInfo.Println("init(): Env vars have been successfully loaded.")
	}
}

// The entry point to the program, creating a bot and connecting to the telegram api,
// launching two goroutines: listening to the telegram api for updates
// and sending a message to the telegram api
func main() {
	// Logging layer
	var log logging.Log
	log.NewLog("logs/bot/")
	defer log.Close()
	var err error

	// Bot object
	var b bot.Bot
	// Creating telegram server connection
	err = b.NewBot()
	if err != nil {
		log.LogFatal.Fatal("main(): Unable to create a new bot, error: ", err)
	} else {
		log.LogInfo.Println("main(): Bot successfully created and connected to telegram api server.")
	}
	defer b.Close()

	// Daemon for listen telegram server chanel
	go b.Listener()

	// Daemon for send our data to telegram server
	err = b.Sender()
	if err != nil {
		log.LogFatal.Fatal("main(): Unable to send messages to telegram server, error: ", err)
	} else {
		log.LogInfo.Println("main(): Sender() has been successfully started.")
	}

	return
}
