package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Z2MMessage represents a message from Zigbee2MQTT
type Z2MMessage struct {
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run capture_messages.go <websocket_url>")
		fmt.Println("Example: go run capture_messages.go ws://localhost:8081/api")
		os.Exit(1)
	}

	url := os.Args[1]
	fmt.Printf("Connecting to %s...\n", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	fmt.Println("Connected! Capturing messages (press Ctrl+C to stop)...")
	fmt.Println(strings.Repeat("=", 60))

	messageCount := 0
	maxMessages := 10

	for messageCount < maxMessages {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		messageCount++
		fmt.Printf("\n--- Message %d ---\n", messageCount)
		fmt.Printf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05.000"))

		// Try to parse as JSON for pretty printing
		var z2mMsg Z2MMessage
		if err := json.Unmarshal(message, &z2mMsg); err == nil {
			prettyJSON, _ := json.MarshalIndent(z2mMsg, "", "  ")
			fmt.Printf("Parsed JSON:\n%s\n", string(prettyJSON))
		} else {
			fmt.Printf("Raw message:\n%s\n", string(message))
		}

		fmt.Println(strings.Repeat("-", 40))
	}

	fmt.Printf("\nCaptured %d messages. Exiting.\n", messageCount)
}
