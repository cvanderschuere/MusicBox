package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"postmaster"
)

const(
	baseURL = "http://www.musicbox.com/"
)

func main() {
	
	//Setup AWS related services (DynamoDB)-defined in aws.go
	if err := setupAWS();err != nil{
		log.Fatal("AWS Login Error: err")
		return
	}
	
	server := postmaster.NewServer()

	//Assign auth callbacks - defined in auth.go
	server.GetAuthSecret = lookupUserSessionID
	server.GetAuthPermissions = getUserPremissions
	server.OnAuthenticated = userConnected
	
	server.MessageToPublish = InterceptMessage //Defined in serverLogic.go
	
	//Setup RPC Functions - defined in rpc.go
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	
	//Unauth rpc
	server.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
		
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: nil}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}



