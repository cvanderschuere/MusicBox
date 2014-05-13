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

	CACHED = true
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

	if CACHED {
		registerCachedRPCs()
	} else {
		registerRPCs()
	}

    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: nil}
	http.Handle("/", s)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func registerRPCs(){
	//Setup RPC Functions - defined in rpc.go
	server.RegisterRPC(baseURL+"userInfo",userInfoRequest)
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	server.RegisterRPC(baseURL+"queue",getQueue)
	server.RegisterRPC(baseURL+"themes",getThemes)
	server.RegisterRPC(baseURL+"trackHistory",getTrackHistory)
	server.RegisterRPC(baseURL+"getNearbyDevices",getNearbyDevices)

	//Unauth rpc
	server.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
	server.RegisterUnauthRPC(baseURL+"trackHistory",getTrackHistory)
	server.RegisterUnauthRPC(baseURL+"getNearbyDevices",getNearbyDevices)
}

func registerCachedRPCs(){
	//Setup RPC Functions - defined in cache.go
	server.RegisterRPC(baseURL+"userInfo",cachedUserInfoRequest)
	server.RegisterRPC(baseURL+"queue",cachedGetQueue)
	server.RegisterRPC(baseURL+"trackHistory",cachedGetTrackHistory)

	server.RegisterUnauthRPC(baseURL+"trackHistory",cachedGetTrackHistory)

	//Setup RPC Functions - defined in rpc.go
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterRPC(baseURL+"themes",getThemes)
	server.RegisterRPC(baseURL+"getNearbyDevices",getNearbyDevices)

	//Unauth rpc
	server.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
	server.RegisterUnauthRPC(baseURL+"getNearbyDevices",getNearbyDevices)


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
