package GetIsChannelLive

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const endPointBase = "https://gql.twitch.tv/gql"

func GetIsChannelLive(clientID, channelName string) []byte {
	client := &http.Client{}
	reqBody := fmt.Sprintf(`[{
		"operationName": "UseLive",
		"variables": {
			"channelLogin": "%s"
		},
		"extensions": {
			"persistedQuery": {
				"version": 1,
				"sha256Hash": "639d5f11bfb8bf3053b424d9ef650d04c4ebb7d94711d644afb08fe9a0fad5d9"
			}
		}
	}]`, channelName)
	req, err := http.NewRequest(
		"POST",
		endPointBase,
		strings.NewReader(reqBody),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("raw: %s", body)
	return body
}
