package main

import (
	"fmt"
	"github.com/cvanderschuere/turnpike"
	"github.com/cvanderschuere/spotify-go"
	"github.com/cvanderschuere/alsa-go"
	"strings"
)

const serverURL = "localhost:8080"

const(
        username string = "christopher.vanderschuere@gmail.com"
        password string = "N0ttingham11"
		deviceName string = "LivingRoom"
)

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
}

/*
	Functions
*/

func main() {
	client := turnpike.NewClient()
	
	//Connect socket between server port and local port
	if err := client.Connect("ws://"+serverURL, "http://localhost:4040"); err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	
	//Make instruction channel
	updateChan := make(chan Notification)
	
	//Launch Event handler
	go eventHandler(client,updateChan)
	
	//Subscribe as appropriate
	client.Subscribe("http://www.MusicBox.com/"+username+"/"+deviceName)
	
	//Login to services & music sink
	controlChan := make(chan bool)
	streamChan := alsa.Init(controlChan)
	
	//Login to spotify (should always work if login test passed)
	ch := spotify.Login(username,password)
	<-ch//Login sync	
	
	
	endOfTrackChan := make(chan bool)//Hack: need chan open initially
	var err error
	
	for{
		select{
		case <-endOfTrackChan:
			//Pass message that track is over
			notiChan <- Notification{Kind:EndOfTrack}
		case update := <-updateChan:
			fmt.Println("Update: ",update.Kind)
			//Take action based on update type
			switch update.Kind{
			case AddedToQueue:
				track := update.Content.(MusicBoxTrack)
				
				//If nothing playing...start it playing
				fmt.Println("Added Track: "+track.Service+" "+track.URL)
			case RemovedFromQueue:
				//Should have to do nothing...unless is current track
				track := update.Content.(MusicBoxTrack)
				
				fmt.Println("Removed Track: "+track.Service+" "+track.URL)
			case PausedTrack:
				//Send pause command				
				fmt.Println("Paused Track")
				controlChan<-false
				
			case ResumedTrack:
				//Send play
				fmt.Println("Resumed Track")
				controlChan<-true
				
			case StoppedTrack:
				//Unload current track
				fmt.Println("Stopped Track")
				spotify.Stop()
				
			case NextTrack:
				//Play track passed
				fmt.Println("Play Next Track")
				track := update.Content.(MusicBoxTrack)
				
				item := &spotify.SpotifyItem{Url:track.URL}
				err,endOfTrackChan = spotify.Play(item,streamChan)
				if err != nil{
					t.Fatalf("Fail: %s",err.Error())
				}
				
			default:
				fmt.Println("Unknown Update Type: %d",update)
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

	//initial queue
	var queue []MusicBoxTrack
	
	EVENT_LOOP:
	for{
		select{
		case event,ok := <-client.HandleChan:
			if ok == false{
				break EVENT_LOOP
			}
			switch event.(type){
			case turnpike.EventMsg:
				commandString := event.(turnpike.EventMsg).Event.(string)
				command := strings.Split(commandString,":")
				
				fmt.Println("Command: "+command[0])
				//Switch through command types
				switch command[0]{
				case "AddTrack":
					//Format: [AddTrack ServiceName TrackName]
					if len(command) >=3{
						newTrack := MusicBoxTrack{Service:command[1],URL:command[2]}
						if queue == nil{
							//create queue
							queue = make([]MusicBoxTrack,1)
							queue[0] = newTrack
							notiChan <- Notification{Kind:AddedToQueue,Content:newTrack} // Start initial playback
						}else{
							//Append
							queue = append(queue,newTrack)
						}
					}
				case "RemoveTrack":
					//Format: [RemoveTrack ServiceName TrackName]
					if len(command) >=3{
						//trackToRemove := MusicBoxTrack{Service:command[1],URL:command[2]}
						
						//Iterate search and remove (Front to back)
						//TODO
						//notiChan <- Notification{Kind:RemovedFromQueue,Content:trackToRemove}
					}
				case "PlayTrack":
					notiChan <- Notification{Kind:ResumedTrack}
				case "PauseTrack":
					notiChan <- Notification{Kind:PausedTrack}
				case "StoppedTrack":
					notiChan <- Notification{Kind:StoppedTrack}
				case "NextTrack":
					if len(queue)>1{
						//Remove current track
						queue = queue[1:]
						
						//Create next track
						next := queue[0]
					
						notiChan <- Notification{Kind:NextTrack,Content:next}
					}
				}
				
			default:
				fmt.Println("Recieved Unknown type")
			}
			
		case update,ok := <-notiChan:
			if ok && update.Kind == EndOfTrack{
				if len(queue)>1{
					//Remove first track
					queue = queue[1:]
				
					//Send update to play next song
					notiChan <- Notification{Kind:NextTrack,Content:queue[0]}
				}
				else{
					//Empty entire list
					queue = nil
				}
				//Publish event
				client.PublishExcludeMe(serverURL+"/"+username+"/"+deviceName,"EndOfTrack")
			}
		}
	}
}