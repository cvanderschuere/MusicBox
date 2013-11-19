package MusicPlayer

import(
	"github.com/cvanderschuere/alsa-go"
	"github.com/cvanderschuere/spotify-go"
	"github.com/jcelliott/lumber"
	"os"
	"os/signal"
	"time"
	"MusicBox/BoxClient/Track"
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

func PlayLoop(updateChan chan Notification, log lumber.Logger, spotifyUsername, spotifyPassword){
	
	//Register for signals from OS to quit if necessary
	signalChan := make(chan os.Signal,1)
	signal.Notify(signalChan)
	
	
	//Login to services & music sink
	controlChan := make(chan bool)
	streamChan := alsa.Init(controlChan)
	
	
	//Login to spotify (should always work if login test passed)
	ch := spotify.Login(spotifyUsername,spotifyPassword)
	<-ch//Login sync	
	
	//Make call for inital songs
	go recommendSongs(4)
	
	
	var endOfTrackChan <-chan bool
	var err error
	
	MAIN_LOOP:
	for{
		select{
		case s := <-signalChan:
			//Recieved signal from OS
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
			case PausedTrack:
				//Send pause command down control CHannel			
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
				// For Logging/Debugging:
				track := update.Content.(Track.Track)
				log.Debug("Play Next Track: "+track.ProviderID)
				
				// Start new Track playing with Spotify Library
				// TODO: definitely make this more generic for multiple music library sources
				item := &spotify.SpotifyItem{Url:track.ProviderID}
				endOfTrackChan,err = spotify.Play(item,streamChan)
				if err != nil{
					log.Error("Error playing track: "+err.Error())
				}
				
			//The Following are used for logging/debugging only:
			case AddedToQueue:
				track := update.Content.(Track.Track)
				log.Debug("Added Track: "+track.ProviderID)
			case RemovedFromQueue:
				track := update.Content.(Track.Track)	
				log.Debug("Removed Track: "+track.ProviderID)
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