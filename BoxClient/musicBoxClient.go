package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/jcelliott/lumber"
	"runtime"
	"os"
	"os/signal"
	"time"
	"MusicBox/BoxClient/Track"
	"MusicBox/BoxClient/MusicPlayer"
	"MusicBox/BoxClient/MusicBoxServer"
)

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"

const musicBoxID = "musicBoxID4"

//Auth info
const WAMP_BASE_URL = "http://api.wamp.ws/"
const WAMP_PROCEDURE_URL = WAMP_BASE_URL+"procedure#"

var authWait = make(chan bool,1) //Used to block until authentication
var boxUsername string
var boxSessionID string

var client *turnpike.Client

var log = lumber.NewConsoleLogger(lumber.TRACE)

const(
        spotifyUsername string = "christopher.vanderschuere@gmail.com"
        spotifyPassword string = "N0ttingham11"
)
var deviceName,_ = os.Hostname()

/*
	Functions
*/

func main() {
	runtime.GOMAXPROCS(2) // Used to regulate main thread managment with libspotify (might not be needed)

	log.Info("Name: "+deviceName+"ID: "+musicBoxID)
	
	
	/*	
	
	//
	// Prepare client
	//	
		
	client = turnpike.NewClient()
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
	config.Header.Add("musicbox-box-id",musicBoxID)


	CONNECT:
		
	if err := client.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		time.Sleep(1)
		goto CONNECT
	}
	
	//
	// Authenticate
	//
	
	//Start session (lookup user & auth)
	client.Call("startSession",baseURL+"musicbox/startSession",musicBoxID)
	
	//Wait until authenticated
	isAuth := <-authWait
	if !isAuth{
		log.Error("Failed auth")
		return
	}
	
	//
	// Connection authenticated
	//
	
	//Launch pinger to keep websocket open (ELB has 60 second timeout)
	go pingClient(client)
	
	//Subscribe as appropriate
	client.Subscribe(baseURL+boxUsername+"/"+musicBoxID)	
	
	*/	
	
	player := MusicPlayer.InitPlayer()
		
	//Make instruction channel
	updateChan := make(chan Player.Notification)
	
	server := MusicBoxServer.InitServer(updateChan)
	
	//Launch Event handler for websocket connection
	go MusicPlayer.EventHandler(log, client, player, updateChan)
	
	// Music Player Loop
	MusicPlayer.PlayLoop(log, updateChan, spotifyUsername, spotifyPassword);
}

func pingClient(client *turnpike.Client){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.PublishExcludeMe(baseURL+"ping","blank")
	}
}

