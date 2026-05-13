package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const endPointBase = "https://gql.twitch.tv/gql"

type twitchResponse struct {
	Data struct {
		Video struct {
			Comments struct {
				Edges []CommentEdge `json:"edges"`
			} `json:"comments"`
		} `json:"video"`
	} `json:"data"`
}

type CommentEdge struct {
	Cursor string `json:"cursor"`
	Node   struct {
		ID        string `json:"id"`
		Commenter struct {
			Login       string `json:"login"`
			DisplayName string `json:"displayName"`
		} `json:"commenter"`
		ContentOffsetSeconds int    `json:"contentOffsetSeconds"`
		CreatedAt            string `json:"createdAt"`
		Message              struct {
			Fragments []struct {
				Text string `json:"text"`
			} `json:"fragments"`
			UserColor  string `json:"userColor"`
			UserBadges []struct {
				SetID   string `json:"setID"`
				Version string `json:"version"`
			} `json:"userBadges"`
		} `json:"message"`
	} `json:"node"`
}

func GetVideoCommentsByOffset(clientId string, videoId string, offset int) {
	currentOffset := offset
	client := &http.Client{}
	var allEdges []CommentEdge
	users := map[string]int{}
	for currentOffset <= 10000 {
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
		]`, videoId, currentOffset)

		req, err := http.NewRequest(
			"POST",
			endPointBase,
			strings.NewReader(reqBody),
		)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Client-ID", clientId)
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
		log.Printf("Status: %d", resp.StatusCode)
		var parsed []twitchResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			log.Fatal(err)
		}
		edges := parsed[0].Data.Video.Comments.Edges
		if len(edges) == 0 {
			break
		}
		for _, edge := range edges {
			user := edge.Node.Commenter.DisplayName
			users[user]++
		}
		allEdges = append(allEdges, edges...)
		currentOffset = allEdges[len(allEdges)-1].Node.ContentOffsetSeconds + 1
	}
	out, err := json.MarshalIndent(struct {
		Comments []CommentEdge  `json:"comments"`
		Users    map[string]int `json:"users"`
	}{Comments: allEdges, Users: users}, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("response.json", out, 0644); err != nil {
		log.Fatal(err)
	}
}
