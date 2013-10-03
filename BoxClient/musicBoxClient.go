package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/jcelliott/lumber"
	"github.com/cvanderschuere/spotify-go"
	"github.com/cvanderschuere/alsa-go"
	"github.com/cvanderschuere/turnpike"
	"runtime"
	"os"
	"os/signal"
	"time"
)

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"

const musicBoxID = "musicBoxID2"

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

type Notification struct{
	Kind NotificationType
	Content interface{}
}

type NotificationType int
const(
	_ NotificationType = iota
	EndOfTrack
	AddedToQueue
	RemovedFromQueue
	PausedTrack
	ResumedTrack
	StoppedTrack
	NextTrack
	//Add more later
)

//Fields must be exported for JSON marshal
type TrackItem struct{
	Title string
	ArtistName string
	AlbumName	string
	ArtworkURL	string
	
	//Track info
	ProviderID	string
	
	//Storage info
	CompositeID	string //username:BoxID
	Date	string  //Date played for accounting purposes
}


/*
	Functions
*/

func main() {
	runtime.GOMAXPROCS(2) // Used to regulate main thread managment with libspotify (might not be needed)

	log.Info("Name: "+deviceName+"ID: "+musicBoxID)
		
	//
	// Prepare client
	//	
		
	client = turnpike.NewClient()
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
	config.Header.Add("musicbox-box-id",musicBoxID)
	
	if err := client.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		return
	}
	
	//Make instruction channel
	updateChan := make(chan Notification)
	
	//Launch Event handler
	go eventHandler(client,updateChan)
	
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
	
	//
	// Prepare music services
	//
	
	//Login to services & music sink
	controlChan := make(chan bool)
	streamChan := alsa.Init(controlChan)
	
	//Login to spotify (should always work if login test passed)
	ch := spotify.Login(spotifyUsername,spotifyPassword)
	<-ch//Login sync	
	
	//Register for signals
	signalChan := make(chan os.Signal,1)
	signal.Notify(signalChan)
	
	//
	// Start main loop
	//
	
	//Make call for inital songs
	go recommendSongs(4)
	
	var endOfTrackChan <-chan bool
	var err error
	
	MAIN_LOOP:
	for{
		select{
		case s := <-signalChan:
			//Recieved signal
			signal.Stop(signalChan)
			log.Debug("Recieved Signal: ", s)
			break MAIN_LOOP
		case <-endOfTrackChan:
			//Pass message that track is over
			log.Trace("Recieved on end of track chan")
			updateChan <- Notification{Kind:EndOfTrack}
			log.Trace("Finished send on end of track update")
		case update := <-updateChan:
			log.Trace("Update: ",update.Kind)
			
			//Take action based on update type
			switch update.Kind{
			case AddedToQueue:
				track := update.Content.(TrackItem)
				log.Debug("Added Track: "+track.ProviderID)
				
			case RemovedFromQueue:
				//Should have to do nothing...unless is current track
				track := update.Content.(TrackItem)	
				log.Debug("Removed Track: "+track.ProviderID)
				
			case PausedTrack:
				//Send pause command				
				log.Debug("Paused Track")
				controlChan<-false
				
			case ResumedTrack:
				//Send play
				log.Debug("Resumed Track")
				controlChan<-true
				
			case StoppedTrack:
				//Unload current track
				log.Debug("Stopped Track")
				spotify.Stop()
				
			case NextTrack:
				//Play track passed
				track := update.Content.(TrackItem)
				log.Debug("Play Next Track: "+track.ProviderID)
				
				//Send startedTrack message
				msg := map[string]interface{} {
					"command":"startedTrack",
					"data": map[string]interface{}{ 
						"deviceID":musicBoxID,
						"track":track,
					},
				}
				client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,msg) //Let others know track has started playing
				
				item := &spotify.SpotifyItem{Url:track.ProviderID}
				endOfTrackChan,err = spotify.Play(item,streamChan)
				if err != nil{
					log.Error("Error playing track: "+err.Error())
				}
				
			default:
				log.Warn("Unknown Update Type: %d",update)
			}	
		}
	}
	
	//
	//Cleanup
	//
	
	//Close alsa stream
	close(streamChan)
	
	//Logout of services
	logout := spotify.Logout()
	<-logout	
}

func pingClient(client *turnpike.Client){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.PublishExcludeMe(baseURL+"ping","blank")
	}
}

