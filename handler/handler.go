package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type Server struct {
	Env      string
	Config   *viper.Viper
	Upgrader websocket.Upgrader
}

const (
	WSExchange = "/ws"
)

func NewServer(
	env string,
	config *viper.Viper,
) (*mux.Router, error) {
	hub := NewHub()
	go hub.run()
	r := mux.NewRouter()
	r.HandleFunc("/ws/{route}", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})
	return r, nil
}
