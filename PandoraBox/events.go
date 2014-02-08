package main

import(
	"github.com/cvanderschuere/turnpike"
	"github.com/cvanderschuere/go-pandora"
	"fmt"
)

var pandoraClient *pandora.PandoraClient

func setupPandora(client * turnpike.Client){
	resp := client.Call(baseURL+"userInfo",musicBoxID)
	message := <-resp
	user := message.Result.(map[string]interface{})
		
	//Use this given info to sign into pandora
	var ch <-chan error
	pandoraClient,ch = pandora.Login(user["Username"].(string),user["PandoraPassword"].(string))
	err := <-ch
	if err != nil{
		fmt.Println(err)
	}
	
	resp = client.Call(baseURL+"boxDetails",[]string{musicBoxID})
	message = <-resp

	boxes := message.Result.(map[string]interface{})		
	thisBox := boxes[musicBoxID].(map[string]interface{})
	boxDetails := thisBox["box"].(map[string]interface{})
	
	playStation(boxDetails["ThemeID"].(string))
}

func handleEvent(topicURI string, event interface{}){
	message := event.(map[string]interface{})
	command := message["command"].(string)
	
	fmt.Println("Command: "+command)
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
		playStation(data["ThemeID"].(string))
		
	default:
		fmt.Println("Unknown message: ",command)
	}
}

func playStation(stationID string){
	fmt.Println(stationID)
	
	station := pandora.Station{ID:stationID}	
	
	ch,_ := pandoraClient.Play(station)
	go func(c <-chan *pandora.Track){
		for track := range c{
			if track == nil{
				continue
			}
			
			//Send this track as started track
			msg := map[string]interface{} {
				"command":"startedTrack",
				"data": map[string]interface{}{ 
					"deviceID":musicBoxID,
					"track":track, //Luckily a TrackItem and pandora.Track are identical :)
				},
			}
			
			client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,msg) //Let others know track has started playing	
		}
	}(ch)
	
	fmt.Println("Finished Playing")
}
