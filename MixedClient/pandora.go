package main

import(
    "github.com/cvanderschuere/go-pandora"
    "fmt"
)


type pandoraClient struct{
    client *pandora.PandoraClient

    wampClient *turnpike.Client
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

    themeId := boxDetails["ThemeID"].(string)

    // Start Station
    pClient.PlayStation(themeId)

    //Set to half initially
    pClient.SetVolume(65)

    return pClient
}


func (c *pandoraClient)SetVolume(vol int){
    c.client.SetVolume(vol)
}

func (c *pandoraClient)Play(){
    c.client.TogglePlayback(true)
}

func (c *pandoraClient)Pause(){
    c.client.TogglePlayback(false)
}

func (c *pandoraClient)NextTrack(){
    c.client.Next()
}

func (c *pandoraClient)PlayStation(stationID string){
    fmt.Println(stationID)

    station := pandora.Station{ID:stationID}

    ch,_ := c.client.Play(station)


    go handlePandoraStart(ch, c.wampClient)

    fmt.Println("Finished Playing")
}


func handlePandoraStart(c <-chan *pandora.Track, client *turnpike.Client){
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
}
