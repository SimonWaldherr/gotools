package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"nhooyr.io/websocket"
)

var (
	rawurl      string
	headerKey   string
	headerValue string
)

func init() {
	flag.StringVar(&rawurl, "url", "", "WebSocket server URL")
	flag.StringVar(&headerKey, "header-key", "", "Custom header key for WebSocket connection")
	flag.StringVar(&headerValue, "header-value", "", "Custom header value for WebSocket connection")
	flag.Parse()
}

func main() {
	_, err := url.Parse(rawurl)
	if err != nil {
		log.Fatal("error parsing: ", err)
	}

	headers := http.Header{}
	if headerKey != "" && headerValue != "" {
		headers.Add(headerKey, headerValue)
	}

	conn, _, err := websocket.Dial(context.Background(), rawurl, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		log.Fatal("error dialing: ", err)
	}
	defer conn.Close(websocket.StatusInternalError, "closing connection")

	go func(conn *websocket.Conn) {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			s := scanner.Text()
			fmt.Println("send=", s)
			err := conn.Write(context.Background(), websocket.MessageText, []byte(s))
			if err != nil {
				log.Fatalf("error sending message: %s", err)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("error scanning stdin: %s", err)
		}
	}(conn)

	for {
		msgType, msg, err := conn.Read(context.Background())
		if err != nil {
			log.Fatal("error receiving: ", err)
		}

		if msgType == websocket.MessageText {
			fmt.Println("recv=", string(msg))
		} else {
			log.Printf("received non-text message type: %v", msgType)
		}
	}
}
