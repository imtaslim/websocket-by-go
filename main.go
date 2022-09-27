package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"websocket/handler"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

const (
	svcName = "codemen-websocket"
	version = "1.0.0"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config := viper.NewWithOptions(
		viper.EnvKeyReplacer(
			strings.NewReplacer(".", "_"),
		),
	)
	config.SetConfigFile("env/config")
	config.SetConfigType("ini")
	config.AutomaticEnv()
	if err := config.ReadInConfig(); err != nil {
		log.Printf("error loading configuration: %v", err)
	}

	s, err := newServer(config)
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", ":"+config.GetString("server.port"))
	if err != nil {
		return err
	}
	log.Printf("Server successfully running on this port: %s", config.GetString("server.port"))

	if err := http.Serve(l, s); err != nil {
		return err
	}
	return nil
}

func newServer(config *viper.Viper) (*mux.Router, error) {
	env := config.GetString("runtime.environment")


	srv, err := handler.NewServer(env, config)
	return srv, err
}