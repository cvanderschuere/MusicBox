package main

import(
	"code.google.com/p/go.net/websocket"
	"github.com/jcelliott/lumber"
	"github.com/cvanderschuere/turnpike"
	"os"
	"os/signal"
	"time"
)

//
// Constants
//

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"
const musicBoxID = "musicBoxID3"

//Auth info
const WAMP_BASE_URL = "http://api.wamp.ws/"
const WAMP_PROCEDURE_URL = WAMP_BASE_URL+"procedure#"
var authWait = make(chan bool,1) //Used to block until authentication
var boxUsername string
var boxSessionID string
var client *turnpike.Client
var log = lumber.NewConsoleLogger(lumber.TRACE)
var deviceName,_ = os.Hostname()

var callback = make(chan bool)

//Fields must be exported for JSON marshal
type TrackItem struct{
	Title string
	ArtistName string
	AlbumName	string
	ArtworkURL	string
	Length	float64
	
	//Track info
	ProviderID	string
	
	//Storage info
	CompositeID	string //username:BoxID
	Date	string  //Date played for accounting purposes
}


func main() {
		
	setupClient()	
		
	//Register for signals
	signalChan := make(chan os.Signal,1)
	signal.Notify(signalChan)
		
	
	//Wait on signal
	s := <-signalChan
	signal.Stop(signalChan)
	log.Debug("Recieved Signal: ", s)	
}

func setupClient()(*turnpike.Client){
	log.Info("Name: "+deviceName+"ID: "+musicBoxID)
		
	//
	// Prepare client
	//	
		
	client := turnpike.NewClient()
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
	config.Header.Add("musicbox-box-id",musicBoxID)


	CONNECT:
		
	if err := client.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		time.Sleep(100*time.Millisecond)
		goto CONNECT
	}
	
	//Launch Event handler
	go eventHandler(client)
	
	//
	// Authenticate
	//
	
	//Start session (lookup user & auth)
	client.Call("startSession",baseURL+"musicbox/startSession",musicBoxID)
	
	//Wait until authenticated
	isAuth := <-authWait
	if !isAuth{
		log.Error("Failed auth")
		return nil
	}
	
	//
	// Connection authenticated
	//
	
	//Launch pinger to keep websocket open (ELB has 60 second timeout)
	go pingClient(client)
	
	client.Call("userInfo",baseURL+"userInfo",musicBoxID)
	<-callback
	
	client.Call("boxDetails",baseURL+"boxDetails",[]string{musicBoxID})
	<-callback
	
	return client
}

func pingClient(client *turnpike.Client){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.PublishExcludeMe(baseURL+"ping","blank")
	}
}