package api

import (

	"fmt"
	"io"
	"net/http"
	"strings"
	"log"
)

const endPointBase  = "https://gql.twitch.tv/gql"

func GetVideoCommentsByOffset(clientId string, videoId string, offset int) {
	client := &http.Client{}

	reqBody := fmt.Sprintf(`[
	{
		"operationName": "VideoCommentsByOffsetOrCursor",
		"variables": {
			"videoID": "%s",
			"contentOffsetSeconds": %d
		},
		"extensions": {
			"persistedQuery": {
				"version": 1,
				"sha256Hash": "b70a3591ff0f4e0313d126c6a1502d79a1c02baebb288227c582044aa76adf6a"
			}
		}
	}
]`, videoId, offset)

	req, err := http.NewRequest(
		"POST",
		endPointBase,
		strings.NewReader(reqBody),
	)
	println(reqBody)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Client-ID", clientId)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Status: %d", resp.StatusCode)
	log.Printf("Response: %s", string(body))
}
