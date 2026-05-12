package main

import (
	"log"
	"net/http"

	"twitchChat/apiGateway"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error loading .env file")
	}

	// clientID := os.Getenv("TWITCH_CLIENT_ID")
	apiGateway.SetupRoutes()
	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
