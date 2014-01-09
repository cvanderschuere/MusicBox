package MusicBoxServer

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"postmaster"
	"MusicBox/BoxClient/MusicPlayer"
)

const(
	baseURL = "http://www.musicbox.com/"
)

type ServerDetails{
	postmaster *postmaster.Server
	updateChan chan MusicPlayer.Notification
}

//Global
var server *ServerDetails

func InitServer(updateChannel chan MusicPlayer.Notification) (*ServerDetails){
	
	//Setup AWS related services (DynamoDB)-defined in aws.go
	if err := setupAWS();err != nil{
		log.Fatal("AWS Login Error: err")
		return
	}
	
	if(server != nil){
		log.Fatal("Error Initializing Server: Server Already Exists")
		return
	}
	
	server.postmaster = postmaster.NewServer()
	server.updateChan = updateChannel

	/*
	//Assign auth callbacks - defined in auth.go
	server.postmaster.GetAuthSecret = lookupUserSessionID
	server.postmaster.GetAuthPermissions = getUserPremissions
	server.postmaster.OnAuthenticated = userConnected
	server.postmaster.OnDisconnect = clientDisconnected
	*/
	
	server.postmaster.MessageToPublish = InterceptMessage //Defined in serverLogic.go
	
	
	//Setup RPC Functions - defined in rpc.go
	server.postmaster.RegisterRPC(baseURL+"players",boxRequest)
	server.postmaster.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.postmaster.RegisterRPC(baseURL+"boxDetails",getMusicBoxDetails)
	server.postmaster.RegisterRPC(baseURL+"trackHistory",getTrackHistory)
	server.postmaster.RegisterRPC(baseURL+"themes",getThemes)
	
	//Unauth rpc
	server.postmaster.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.postmaster.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
		
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: nil}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	
	return &server
}



