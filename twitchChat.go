package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"twitchChat/api"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error loading .env file")
	}

	clientID := os.Getenv("TWITCH_CLIENT_ID")

	api.GetVideoCommentsByOffset(clientID, "2754561063", 60)
}
