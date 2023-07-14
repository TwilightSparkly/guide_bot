.PHONY:
.SILENT:

include cfg/.env

buildBot:
	go build -o ./.bin/bot cmd/bot/main.go

runBot: buildBot
	./.bin/bot

buildAi:
	go build -o ./.bin/ai cmd/ai/main.go

runAi: buildAi
	./.bin/ai