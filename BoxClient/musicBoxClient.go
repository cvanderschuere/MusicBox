package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/cvanderschuere/turnpike"
	"github.com/jcelliott/lumber"
	"github.com/cvanderschuere/spotify-go"
	"github.com/cvanderschuere/alsa-go"
	"runtime"
	"os"
	"os/signal"
	"time"
)

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"

const musicBoxID = "musicBoxID2"

var client *turnpike.Client

var log = lumber.NewConsoleLogger(lumber.TRACE)

const(
        username string = "christopher.vanderschuere@gmail.com"
        password string = "N0ttingham11"
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
	
	//Launch pinger to keep websocket open (ELB has 60 second timeout)
	go pingClient(client)
	
	//Subscribe as appropriate
	client.Subscribe(baseURL+username+"/"+deviceName)
	client.Subscribe(baseURL+username+"/"+deviceName+"/internal") //Also recieve music box exclusive events
	
	//
	// Prepare music services
	//
	
	//Login to services & music sink
	controlChan := make(chan bool)
	streamChan := alsa.Init(controlChan)
	
	//Login to spotify (should always work if login test passed)
	ch := spotify.Login(username,password)
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
				client.PublishExcludeMe(baseURL+username+"/"+deviceName,msg) //Let others know track has started playing
				
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
	
	//Cleanup
	
	//Close alsa stream
	close(streamChan)
	
	//Logout of services
	logout := spotify.Logout()
	<-logout	
}

//Decoded event into music box instruction
//This is the only function allowed to add/remove from the upcoming queue
func eventHandler(client *turnpike.Client, notiChan chan Notification){

	//initial queue...maybe fetch update from server with rpc call
	var queue []TrackItem
	var isPlaying bool = false
	
	EVENT_LOOP:
	for{
		log.Trace("Event Handler Select")
		select{
		case event,ok := <-client.HandleChan:
			if ok == false{
				break EVENT_LOOP
			}
			switch event.(type){
			case turnpike.EventMsg:
				message := event.(turnpike.EventMsg).Event.(map[string]interface{})
				
				log.Debug("Command: "+ message["command"].(string))
				//Switch through command types
				switch message["command"]{
				case "addTrack":
					data := message["data"].([]interface{}) // Need for interface due to interal marshaling in turnpike
					
					//Add all passed tracks
					for _,trackDict := range data {
						track := trackDict.(map[string]interface{})
						newTrack := TrackItem{ProviderID:track["ProviderID"].(string),Title:track["Title"].(string),ArtistName:track["ArtistName"].(string),AlbumName:track["albumName"].(string)}
						if queue == nil{
							//create queue
							queue = make([]TrackItem,1)
							queue[0] = newTrack
							notiChan <- Notification{Kind:NextTrack,Content:newTrack} // Start initial playback
							isPlaying = true
						
							playMsg := map[string]string{
								"command":"playTrack",
							}
							client.PublishExcludeMe(baseURL+username+"/"+deviceName,playMsg) //Let others know track is playing
						}else{
							//Append
							queue = append(queue,newTrack)
							notiChan <- Notification{Kind:AddedToQueue,Content:newTrack} // Give chance to preload
						}
					}
					
					//Queue must add recommendation to stay at minimum 2
					if len(queue) == 1{
						log.Trace("Finding similar songs to add")
						go recommendSongs(3)
					}
					
					
				case "removeTrack":
					//Format: [RemoveTrack ServiceName TrackName]
						//trackToRemove := MusicBoxTrack{Service:command[1],URL:command[2]}
						
					//Iterate search and remove (Front to back)
					//TODO
					//notiChan <- Notification{Kind:RemovedFromQueue,Content:trackToRemove}
				case "playTrack":
					isPlaying = true
					notiChan <- Notification{Kind:ResumedTrack}
				case "pauseTrack":
					isPlaying = false
					notiChan <- Notification{Kind:PausedTrack}
				case "stopTrack":
					isPlaying = false
					notiChan <- Notification{Kind:StoppedTrack} //Song stays in queue...no different than pause?
				case "nextTrack":
					if len(queue)>1{
						//Remove current track
						queue = queue[1:]
						
						//Create next track
						next := queue[0]
					
						isPlaying = true	
						notiChan <- Notification{Kind:NextTrack,Content:next}
						
						//Make sure queue has enough recommendations
						if len(queue) <= 1{
							log.Trace("Finding similar songs to add")
							go recommendSongs(1)
						}
					}
				//
				//Internal Events
				//
				case "QueueRequest":
					//Publish queue update...only music box responds to this but all client should recieve CurrentQueue
					client.PublishExcludeMe(baseURL+username+"/"+deviceName,queue)
				case "statusUpdate":
					//Send back map of current status values: title,isPlaying,queue
					response := map[string]interface{}{
						"deviceName": deviceName,
						"isPlaying": isPlaying,
						"queue": queue,	
					}
					
					responseMessage := map[string]interface{}{
						"command": "statusUpdate",
						"data": response,
					}
					
					client.PublishExcludeMe(baseURL+username+"/"+deviceName,responseMessage)
				}
				
			case turnpike.CallResultMsg:
				message := event.(turnpike.CallResultMsg)
				if message.CallID == "recommendSongs"{
					tracks := message.Result.([]interface{})
					
					log.Info("Adding %d recommendations to queue",len(tracks))
					
					for _,m := range tracks{
						track := m.(map[string]interface{})
						t := TrackItem{ProviderID:track["ProviderID"].(string),Title:track["Title"].(string),ArtistName:track["ArtistName"].(string),AlbumName:track["AlbumName"].(string),ArtworkURL:track["ArtworkURL"].(string)}
						
						queue = append(queue,t)					
					}	
					
					if !isPlaying && len(queue) > 0{
						notiChan <- Notification{Kind:NextTrack,Content:queue[0]} // Start initial playback
					}			
			}				
				
			default:
				log.Warn("Recieved Unknown type")
			}
			
		case update,ok := <-notiChan:
			if ok && update.Kind == EndOfTrack{
				if len(queue)>1{
					//Remove first track
					queue = queue[1:]
				
					log.Trace("Moving to next song")
					isPlaying = true
					//Send update to play next song
					notiChan <- Notification{Kind:NextTrack,Content:queue[0]}
				}else{
					log.Trace("Clear queue")
					//Empty entire list
					queue = nil
					isPlaying = false
				}
				
				if len(queue) < 2{
					go recommendSongs(3) //Add radio never ending playlist
				}
				
				//Publish event
				msg := map[string]string{
					"command":"endOfTrack",
				}
				
				client.PublishExcludeMe(baseURL+username+"/"+deviceName,msg)
			}
		}
	}
}

func pingClient(client *turnpike.Client){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.PublishExcludeMe(baseURL+"ping","blank")
	}
}

