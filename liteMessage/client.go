// Description: A simple client for the liteMessage server.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Message struct {
	Type    string `json:"type"`
	Topic   string `json:"topic,omitempty"`
	Payload string `json:"payload,omitempty"`
}

func handleMessage(msg *Message) {
	fmt.Printf("[Topic: %s] %s\n", msg.Topic, msg.Payload)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8083")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return
	}
	defer conn.Close()

	// Authenticate if necessary
	// Send an authentication message with the format "username:password"
	// _, _ = conn.Write([]byte(`{"type": "auth", "payload": "user:pass"}` + "\n"))

	scanner := bufio.NewScanner(conn)
	go func() {
		for scanner.Scan() {
			var msg Message
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				fmt.Printf("Error unmarshalling message: %v\n", err)
				continue
			}

			handleMessage(&msg)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.SplitN(input, " ", 2)

		if len(parts) == 2 {
			cmd := parts[0]
			payload := parts[1]

			switch cmd {
			case "subscribe":
				topic := payload
				fmt.Printf("Subscribing to topic '%s'\n", topic)
				_, _ = conn.Write([]byte(fmt.Sprintf(`{"type": "subscribe", "topic": "%s"}`+"\n", topic)))
			case "publish":
				parts = strings.SplitN(payload, " ", 2)
				if len(parts) == 2 {
					topic := parts[0]
					message := parts[1]
					fmt.Printf("Publishing message to topic '%s': %s\n", topic, message)
					_, _ = conn.Write([]byte(fmt.Sprintf(`{"type": "publish", "topic": "%s", "payload": "%s"}`+"\n", topic, message)))
				}
			default:
				fmt.Println("Unknown command. Use 'subscribe <topic>' or 'publish <topic> <message>'.")
			}
		} else {
			fmt.Println("Invalid input. Use 'subscribe <topic>' or 'publish <topic> <message>'.")
		}
	}
}
