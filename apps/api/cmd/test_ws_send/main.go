package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	token := flag.String("token", "", "Bearer token")
	toUser := flag.String("to", "", "To User ID")
	msg := flag.String("msg", "", "Message content")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "token=" + *token}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	payload := map[string]string{
		"type":     "chat",
		"toUserId": *toUser,
		"content":  *msg,
	}
	bytes, _ := json.Marshal(payload)

	if err := c.WriteMessage(websocket.TextMessage, bytes); err != nil {
		log.Fatal("write:", err)
	}

	// Wait a bit for server to process
	time.Sleep(time.Second)
}
