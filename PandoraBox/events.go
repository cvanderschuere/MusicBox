package main

import(
	"github.com/cvanderschuere/turnpike"
	"github.com/cvanderschuere/go-pandora"
	"fmt"
)

var pandoraClient *pandora.PandoraClient

//Decoded event into music box instruction
//This is the only function allowed to add/remove from the upcoming queue
func eventHandler(client *turnpike.Client){
	
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
				handleEvent(event.(turnpike.EventMsg))
					
			case turnpike.CallResultMsg:
				handleResult(event.(turnpike.CallResultMsg))
				
			default:
				log.Warn("Recieved Unknown type")
			}
		}
	}
}

func handleEvent(event turnpike.EventMsg){
	message := event.Event.(map[string]interface{})
	command := message["command"].(string)
	
	switch command{
	case "playTrack":
		pandoraClient.TogglePlayback(true)
	case "pauseTrack":
		pandoraClient.TogglePlayback(false)
	case "nextTrack":
		pandoraClient.Next()
	case "updateTheme":
		//Extract station ID
		data := message["data"].(map[string]interface{})
		station := pandora.Station{ID:data["ID"].(string)}

		pandoraClient.Play(station)
	default:
		fmt.Println("Unknown message: ",command)
	}
}

func handleResult(message turnpike.CallResultMsg){
	fmt.Println(message.CallID)

	switch message.CallID {
	case "userInfo":
		user := message.Result.(map[string]interface{})
			
		//Use this given info to sign into pandora
		pandoraClient,_ = pandora.Login(user["Username"].(string),user["PandoraPassword"].(string))
		callback<-true
	
	case "boxDetails":
		boxes := message.Result.(map[string]interface{})
		thisBox := boxes[musicBoxID].(map[string]interface{})
		
		pandoraClient.Play(thisBox["Theme"].(string))
		
		callback<-true
		
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
	}
}