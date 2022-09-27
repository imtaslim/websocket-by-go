package handler

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWs(t *testing.T) {
	wsURL := "ws://localhost:8888/ws/order"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	// Write a message.
	c.WriteMessage(websocket.TextMessage, []byte("hello"))

	// Expect the server to echo the message back.
	c.SetReadDeadline(time.Now().Add(time.Second * 2))
	mt, msg, err := c.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if mt != websocket.TextMessage || string(msg) != "hello" {
		t.Errorf("expected text hello, got %d: %s", mt, msg)
	}
}
