package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/cvanderschuere/turnpike"
	"io/ioutil"
	"log"
	"net/http"
	"postmaster"
)

const(
	baseURL = "http://www.musicbox.com/"
)

//Global
var server *postmaster.Server

func main() {
	go startWebServer()

	
	//Setup AWS related services (DynamoDB)-defined in aws.go
	if err := setupAWS();err != nil{
		log.Fatal("AWS Login Error: err")
		return
	}
	
	server = postmaster.NewServer()

	//Assign auth callbacks - defined in auth.go
	server.GetAuthSecret = lookupUserSessionID
	server.GetAuthPermissions = getUserPremissions
	server.OnAuthenticated = userConnected
	server.OnDisconnect = clientDisconnected
	
	server.MessageToPublish = InterceptMessage //Defined in serverLogic.go
	
	//Setup RPC Functions - defined in rpc.go
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	server.RegisterRPC(baseURL+"trackHistory",getTrackHistory)
	server.RegisterRPC(baseURL+"themes",getThemes)
	
	//Unauth rpc
	server.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
		
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: nil}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func startWebServer() {
	webServer := http.NewServeMux()

	webServer.HandleFunc("/", serveHomePage)

	if err := http.ListenAndServe(":80", webServer); err != nil {
		log.Fatal("Unable to Start Web Server: ", err)
	}
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile("/home/ubuntu/MusicBoxWebClient/index.html")

	if err != nil {
		return
	}

	w.Write([]byte(body))
}
