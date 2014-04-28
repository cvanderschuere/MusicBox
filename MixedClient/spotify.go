package main

import(
    "github.com/cvanderschuere/turnpike"
    "github.com/cvanderschuere/spotify-go"
    "github.com/cvanderschuere/alsa-go"
    "fmt"
)

type spotifyClient struct{
    controlChan chan bool
    streamChan chan<- alsa.AudioStream
}

func SetupSpotify(client *turnpike.Client) *spotifyClient{
    sClient := new(spotifyClient)

    sClient.controlChan = make(chan bool)
    sClient.streamChan = alsa.Init(sClient.controlChan)

    //Login to spotify (should always work if login test passed)
    ch := spotify.Login(spotifyUsername,spotifyPassword)
    <-ch// Wait for login sync

    return sClient
}

func (c *spotifyClient)Play(){
    c.controlChan<-true
}

func (c *spotifyClient)Pause(){
    c.controlChan<-false
}

func (c *spotifyClient)Stop(){
    spotify.Stop()
}


func (c *spotifyClient)NextTrack(track TrackItem, endOfTrackChan *<-chan bool) (<-chan bool){

    log.Debug("channel: ", endOfTrackChan)
    log.Debug("channelAddr: ", &endOfTrackChan)

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

    var err error
    endOfTrackChan,err = spotify.Play(item, c.streamChan)

    if err != nil{
        fmt.Println("Error playing track: "+err.Error())
    }

    log.Debug("channel: ", endOfTrackChan)
    log.Debug("channelAddr: ", &endOfTrackChan)

    return endOfTrackChan
}

func (c *spotifyClient)ShutDown(){
    //Close alsa stream
    close(c.streamChan)

    //Logout of services
    logout := spotify.Logout()
    <-logout
}
