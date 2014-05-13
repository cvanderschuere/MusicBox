package main

import(
    "github.com/cvanderschuere/turnpike"
    "github.com/cvanderschuere/go-pandora"
    "fmt"
)


type pandoraClient struct{
    client *pandora.PandoraClient
	ThemeID string //Current themeID

    trackChan <-chan *pandora.Track
}

func SetupPandora(client *turnpike.Client) (*pandoraClient){

    pClient := new(pandoraClient)

    // Request pandora info
    resp := client.Call(baseURL+"userInfo",musicBoxID)
    message := <-resp
    user := message.Result.(map[string]interface{})

    var ch <-chan error
    pClient.client,ch = pandora.Login(user["Username"].(string),user["PandoraPassword"].(string))

    err := <-ch
    if err != nil{
        fmt.Println(err)
    }

    // Get Box Details so we can play the current station
    resp = client.Call(baseURL+"boxDetails",[]string{musicBoxID})
    message = <-resp

    boxes := message.Result.(map[string]interface{})
    thisBox := boxes[musicBoxID].(map[string]interface{})
    boxDetails := thisBox["box"].(map[string]interface{})

	// Save theme
    pClient.ThemeID = boxDetails["ThemeID"].(string)

    pClient.PlayStation(pClient.ThemeID)

    //Set to half initially
    pClient.SetVolume(65)

    return pClient
}


func (c *pandoraClient)SetVolume(vol uint8){
    c.client.SetVolume(vol)
}

func (c *pandoraClient)Play(){
    c.client.TogglePlayback(true)
}

func (c *pandoraClient)Pause(){
    c.client.TogglePlayback(false)
}

func (c *pandoraClient)Stop(){
    c.client.Stop()

    c.trackChan = make(chan *pandora.Track,0)
}

func (c *pandoraClient)NextTrack(){
    c.client.Next()
}

func (c *pandoraClient)PlayStation(stationID string){
    fmt.Println(stationID)

    c.ThemeID = stationID

    station := pandora.Station{ID:stationID}

    ch,_ := c.client.Play(station)

	go func(ch <-chan *pandora.Track){
	for track := range ch{
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

    c.trackChan = ch

}
