package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"github.com/cvanderschuere/turnpike"
	//"turnpike"
)

//Global
var server *turnpike.Server

func main() {
	server = turnpike.NewServer()
	
	//Setup RPC Functions...wrong place to do this
	//We should be doing this per non-music box client
	server.RegsterRPC("http://www.MusicBox.com/christopher.vanderschuere@gmail.com/LivingRoom/currentQueueRequest",queueRequest)
	
	http.Handle("/", websocket.Handler(turnpike.HandleWebsocket(server)))
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

//RPC Handler of form: res, err = f(id, msg.ProcURI, msg.CallArgs...)
func queueRequest(id, url string, args ...interface{})(interface{},error){
	//Recieved request for queue...for now just pass on to music box
	
	//This will forward an event on a private channel to the music box
	//The music box will then publish a typical CurrentQueue update to everyone
	server.SendEvent(url+"/internal","QueueRequest");
	
	return "QueueRequest Sent",nil
}