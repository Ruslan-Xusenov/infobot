package main

import (
	"log"

	"github.com/company/infobot/bot"
	"github.com/company/infobot/config"
	"github.com/company/infobot/database"
)

func main() {
	cfg := config.Load()

	err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	bot.Start(cfg)
}
