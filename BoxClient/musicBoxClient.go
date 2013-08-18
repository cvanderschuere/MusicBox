package main

import (
	"github.com/cvanderschuere/turnpike"
	"github.com/jcelliott/lumber"
	"github.com/cvanderschuere/spotify-go"
	"github.com/cvanderschuere/alsa-go"
	"runtime"
	"os"
)

const serverURL = "ec2-54-218-97-11.us-west-2.compute.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"

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

type MusicBoxTrack struct{
	Service string
	URL		string
	AlbumName string
	ArtistName string
	TrackName	string
}

/*
	Functions
*/

func main() {
	runtime.GOMAXPROCS(2) // Used to regulate main thread managment with libspotify (might not be needed)
	
	client := turnpike.NewClient()
	
	//Connect socket between server port and local port
	if err := client.Connect("ws://"+serverURL, "http://localhost:4040"); err != nil {
		log.Error("Error connecting: ", err)
		return
	}
	
	//Make instruction channel
	updateChan := make(chan Notification)
	
	//Launch Event handler
	go eventHandler(client,updateChan)
	
	//Subscribe as appropriate
	client.Subscribe(baseURL+username+"/"+deviceName)
	client.Subscribe(baseURL+username+"/"+deviceName+"/internal") //Also recieve music box exclusive events
	
	//Login to services & music sink
	controlChan := make(chan bool)
	streamChan := alsa.Init(controlChan)
	
	//Login to spotify (should always work if login test passed)
	ch := spotify.Login(username,password)
	<-ch//Login sync	
	
	var endOfTrackChan <-chan bool
	var err error
	
	for{
		select{
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
				track := update.Content.(MusicBoxTrack)
				
				//If nothing playing...start it playing
				log.Debug("Added Track: "+track.Service+" "+track.URL)
			case RemovedFromQueue:
				//Should have to do nothing...unless is current track
				track := update.Content.(MusicBoxTrack)
				
				log.Debug("Removed Track: "+track.Service+" "+track.URL)
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
				track := update.Content.(MusicBoxTrack)
				log.Debug("Play Next Track: "+track.URL)
				
				item := &spotify.SpotifyItem{Url:track.URL}
				err,endOfTrackChan = spotify.Play(item,streamChan)
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
	var queue []MusicBoxTrack
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
					data := message["data"].([]interface{})
					
					for _,trackDict := range data {
						track := trackDict.(map[string]interface{})
						newTrack := MusicBoxTrack{Service:track["service"].(string),URL:track["url"].(string),TrackName:track["trackName"].(string),ArtistName:track["artistName"].(string),AlbumName:track["albumName"].(string)}
						if queue == nil{
							//create queue
							queue = make([]MusicBoxTrack,1)
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
					}
					//Else don't allow
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
				//Publish event
				msg := map[string]string{
					"command":"endOfTrack",
				}
				
				client.PublishExcludeMe(baseURL+username+"/"+deviceName,msg)
			}
		}
	}
}
