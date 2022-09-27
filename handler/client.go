package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	userID string
	url    string
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}
type ExchangeData struct {
	UserID  string
	URL     string
	Payload []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		generateData := ExchangeData{
			UserID:  c.userID,
			URL:     c.url,
			Payload: message,
		}

		handleSocketPayloadEvents(c, generateData)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	params := mux.Vars(r)
	route := params["route"]
	userID := GetUserIdFormRoute(route)

	client := &Client{userID: userID, url: route, hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func handleSocketPayloadEvents(client *Client, exchangeData ExchangeData) {
	switch {
	case exchangeData.UserID != "":
		EmitToSpecificClient(client.hub, exchangeData)
	default:
		BroadcastSocketEventToAllClient(client.hub, exchangeData)
	}
}

func EmitToSpecificClient(hub *Hub, ed ExchangeData) {
	for client := range hub.clients {
		if client.userID == ed.UserID && client.url == ed.URL {
			select {
			case client.send <- ed.Payload:
			default:
				close(client.send)
				delete(hub.clients, client)
			}
		}
	}
}

func BroadcastSocketEventToAllClient(hub *Hub, ed ExchangeData) {
	for client := range hub.clients {
		if client.url == ed.URL {
			select {
			case client.send <- ed.Payload:
			default:
				close(client.send)
				delete(hub.clients, client)
			}
		}
	}
}

func GetUserIdFormRoute(route string) string {
	for _, item := range strings.Split(route, "=") {
		_, err := uuid.Parse(item)
		if err == nil {
			return item
		}
	}

	return ""
}
