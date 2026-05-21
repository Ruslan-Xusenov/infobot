package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	AdminID    int64
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error reading it, using OS env vars")
	}

	adminIDStr := os.Getenv("ADMIN_ID")
	adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)

	return &Config{
		BotToken:   os.Getenv("BOT_TOKEN"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		AdminID:    adminID,
	}
}
