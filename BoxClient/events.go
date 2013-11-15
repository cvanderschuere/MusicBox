package main

import(
	"github.com/cvanderschuere/turnpike"
	"MusicBox/BoxClient/Track"
)

//Decoded event into music box instruction
//This is the only function allowed to add/remove from the upcoming queue
func eventHandler(client *turnpike.Client, notiChan chan Notification){

	//initial queue...maybe fetch update from server with rpc call
	var queue []Track.Track
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
						newTrack := Track.FromMap(track)
						if queue == nil{
							//create queue
							queue = make([]Track.Track,1)
							queue[0] = newTrack
							notiChan <- Notification{Kind:NextTrack,Content:newTrack} // Start initial playback
							isPlaying = true
						
							playMsg := map[string]string{
								"command":"playTrack",
							}
							client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,playMsg) //Let others know track is playing
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
				case "updateTheme":
					go recommendSongs(1)
					queue = []Track.Track{}
					
				//
				//Internal Events
				//
				case "QueueRequest":
					//Publish queue update...only music box responds to this but all client should recieve CurrentQueue
					client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,queue)
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
					
					client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,responseMessage)
				}
				
			case turnpike.CallResultMsg:
				message := event.(turnpike.CallResultMsg)
				
				switch message.CallID {
				case "recommendSongs":
					tracks := message.Result.([]interface{})
					
					log.Info("Adding %d recommendations to queue",len(tracks))
					
					for _,m := range tracks{
						track := m.(map[string]interface{})
						t := Track.FromMap(track)
						queue = append(queue,t)					
					}	
					
					if !isPlaying && len(queue) > 0{
						notiChan <- Notification{Kind:NextTrack,Content:queue[0]} // Start initial playback
					}	
				case "startSession":
					user := message.Result.(map[string]interface{})
					boxUsername = user["username"].(string)
					boxSessionID = user["sessionID"].(string)
					
					extra := map[string]interface{}{
						"client-type":"musicBox-v1", //Used to diferentiate musicbox from other clients (ie Website)
						"client-id":musicBoxID,
					}
	
					client.Call("authreq",WAMP_PROCEDURE_URL+"authreq",boxUsername,extra)
				case "authreq":
					//Recieve challenge
					ch,ok := message.Result.(string)
					if !ok{
						log.Error("Incorrect response type")
					}
					
					//Calculate & send signature
					sig := authSignature([]byte(ch),boxSessionID,nil)
					client.Call("auth",WAMP_PROCEDURE_URL+"auth",sig)
					
				case "auth":
					//Recieve permission information
					authWait<-true
						
			default:
				log.Warn("Recieved Unknown type")
			}
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
				
				client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,msg)
			}
		}
	}
}
