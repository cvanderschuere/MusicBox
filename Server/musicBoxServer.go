package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"github.com/cvanderschuere/turnpike"
	//"turnpike"
)

func main() {
	server := turnpike.NewServer()
	
	http.Handle("/", websocket.Handler(turnpike.HandleWebsocket(server)))
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}