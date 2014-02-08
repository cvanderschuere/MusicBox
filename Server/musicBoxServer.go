package main

import (
	"code.google.com/p/go.net/websocket"
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
		log.Fatal("AWS Login Error:",err)
	}
	
	server = postmaster.NewServer()

	//Assign auth callbacks - defined in auth.go
	server.GetAuthSecret = lookupUserSessionID
	server.GetAuthPermissions = getUserPremissions
	server.OnAuthenticated = userConnected
	server.OnDisconnect = clientDisconnected
	
	server.MessageToPublish = InterceptMessage //Defined in serverLogic.go
	
	//Setup RPC Functions - defined in rpc.go
	server.RegisterRPC(baseURL+"userInfo",userInfoRequest)
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	server.RegisterRPC(baseURL+"trackHistory",getTrackHistory)
	server.RegisterRPC(baseURL+"queue",getQueue)
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

	webServer.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("/home/ubuntu/MusicBoxWebClient/css"))))
	webServer.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("/home/ubuntu/MusicBoxWebClient/js"))))
	webServer.Handle("/font/", http.StripPrefix("/font/", http.FileServer(http.Dir("/home/ubuntu/MusicBoxWebClient/font"))))
	webServer.Handle("/template/", http.StripPrefix("/template/", http.FileServer(http.Dir("/home/ubuntu/MusicBoxWebClient/template"))))
	
	if err := http.ListenAndServe(":2020", webServer); err != nil {
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
