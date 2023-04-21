// Description: A simple message broker written in Go
// It supports topic-based subscriptions.
// The protocol is very simple and uses JSON for messages.
// The server can be started with the following command:
// go run server.go -addr :8083 -auth -username user -password pass
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Message struct {
	Type    string `json:"type"`
	Topic   string `json:"topic,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type Client struct {
	conn       net.Conn
	id         string
	subscribed map[string]bool
	auth       bool
}

type Server struct {
	clients      map[string]*Client
	clientsMutex sync.Mutex
	authEnabled  bool
	username     string
	password     string
}

func NewServer(authEnabled bool, username, password string) *Server {
	return &Server{
		clients:     make(map[string]*Client),
		authEnabled: authEnabled,
		username:    username,
		password:    password,
	}
}

func (s *Server) handleClientMessage(client *Client, msg *Message) {
	switch msg.Type {
	case "auth":
		if s.authEnabled && msg.Payload == fmt.Sprintf("%s:%s", s.username, s.password) {
			client.auth = true
			_, _ = client.conn.Write([]byte(`{"type":"auth", "payload":"success"}` + "\n"))
		} else {
			_, _ = client.conn.Write([]byte(`{"type":"auth", "payload":"failed"}` + "\n"))
		}
	case "subscribe":
		if client.auth || !s.authEnabled {
			client.subscribed[msg.Topic] = true
		}
	case "publish":
		if client.auth || !s.authEnabled {
			s.clientsMutex.Lock()
			defer s.clientsMutex.Unlock()

			data := []byte(fmt.Sprintf(`{"type":"publish", "topic":"%s", "payload":"%s"}`+"\n", msg.Topic, msg.Payload))

			for _, c := range s.clients {
				if c.id != client.id && c.subscribed[msg.Topic] {
					go func(receiver *Client) {
						if _, err := receiver.conn.Write(data); err != nil {
							log.Printf("Error forwarding message to client %s: %v", receiver.id, err)
						}
					}(c)
				}
			}
		}
	}
}

func (s *Server) handleClientConnection(client *Client) {
	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, client.id)
		s.clientsMutex.Unlock()
		client.conn.Close()
	}()

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			log.Printf("Error unmarshalling message from client %s: %v", client.id, err)
			continue
		}

		s.handleClientMessage(client, &msg)
	}
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("LiteMessage server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		id := fmt.Sprintf("%s", conn.RemoteAddr())
		log.Printf("Client connected: %s", id)

		client := &Client{
			conn:       conn,
			id:         id,
			subscribed: make(map[string]bool),
			auth:       false,
		}

		s.clientsMutex.Lock()
		s.clients[id] = client
		s.clientsMutex.Unlock()

		go s.handleClientConnection(client)
	}
}

func main() {
	var addr string
	var authEnabled bool
	var username string
	var password string

	flag.StringVar(&addr, "addr", "localhost:8083", "Server address")
	flag.BoolVar(&authEnabled, "auth", false, "Enable authentication")
	flag.StringVar(&username, "user", "user", "Username for authentication")
	flag.StringVar(&password, "pass", "pass", "Password for authentication")

	flag.Parse()

	server := NewServer(authEnabled, username, password)

	if err := server.Start(addr); err != nil {
		log.Printf("Error starting server: %v", err)
		os.Exit(1)
	}
}
