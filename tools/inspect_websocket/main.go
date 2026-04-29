package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

func main() {
	url := flag.String("url", "ws://localhost:8081/api", "Zigbee2MQTT WebSocket URL")
	topic := flag.String("topic", "", "Filter by topic (substring match)")
	field := flag.String("field", "", "Filter to messages containing this top-level JSON field")
	limit := flag.Int("n", 0, "Stop after N matching messages (0 = unlimited)")
	initialOnly := flag.Bool("initial", false, "Stop after initial burst (no new messages for 2s)")
	ota := flag.Bool("ota", false, "Shorthand: show only messages with an 'update' field")
	flag.Parse()

	if *ota {
		f := "update"
		field = &f
	}

	conn, _, err := websocket.DefaultDialer.Dial(*url, nil)
	if err != nil {
		log.Fatal("connect:", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("close: %v", cerr)
		}
	}()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", *url)

	matched := 0
	lastReceived := time.Now()

	for {
		if *initialOnly {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second)) //nolint:errcheck
		}

		_, raw, err := conn.ReadMessage()
		if err != nil {
			if *initialOnly && time.Since(lastReceived) >= 2*time.Second {
				break
			}
			log.Fatal("read:", err)
		}
		lastReceived = time.Now()

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		if *topic != "" && !strings.Contains(msg.Topic, *topic) {
			continue
		}

		if *field != "" {
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				continue
			}
			if _, has := payload[*field]; !has {
				continue
			}
		}

		out, _ := json.MarshalIndent(msg, "", "  ")
		fmt.Printf("%s\n", out)
		matched++

		if *limit > 0 && matched >= *limit {
			break
		}
	}

	fmt.Fprintf(os.Stderr, "Matched %d message(s).\n", matched)
}
