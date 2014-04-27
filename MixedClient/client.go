package main

import(
    "code.google.com/p/go.net/websocket"
    "github.com/jcelliott/lumber"
    "github.com/cvanderschuere/turnpike"
    "os"
    "os/signal"
    "time"
)


const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"
const musicBoxID = "musicBoxID1"

const spotifyUsername string = "christopher.vanderschuere@gmail.com"
const spotifyPassword string = "N0ttingham11"

var boxUsername string
var boxSessionID string


var client *turnpike.Client

var pandoraPlaying bool = true

var log = lumber.NewConsoleLogger(lumber.TRACE)


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

    ChangeTheme
    SetVolume
    StatusUpdate
    //Add more later
)

//Fields must be exported for JSON marshal
type TrackItem struct{
    Title string
    ArtistName string
    AlbumName	string
    ArtworkURL	string
    Length	float64

    //Track info
    ProviderID	string

    //Storage info
    CompositeID	string //username:BoxID
    Date	string  //Date played for accounting purposes
}

func trackItemFromMap(data map[string]interface{})(TrackItem){

    //Extract length
    //l,_ := strconv.ParseFloat(data["Length"].(string),32)

    t := TrackItem{
        Title: data["Title"].(string),
        ArtistName: data["ArtistName"].(string),
        AlbumName: data["AlbumName"].(string),
        ArtworkURL: data["ArtworkURL"].(string),
        ProviderID: data["ProviderID"].(string),
        Length: data["Length"].(float64),
        Date: data["Date"].(string),
    }

    return t
}





var updateChan chan Notification

func main(){

    pClient, sClient := initializeClient()

    //Register for signals
    signalChan := make(chan os.Signal,1)
    signal.Notify(signalChan)

    playerLoop(signalChan, pClient, sClient)

    sClient.ShutDown()
}


func playerLoop(signalChan chan os.Signal, pClient *pandoraClient, sClient *spotifyClient){

    spotifyEndChan := make(<-chan bool)
    var queue []TrackItem = make([]TrackItem,0)
    var isPlaying = true

    delayedAction := new(Notification)

LOOP:
    select{
    case s := <- signalChan:
        signal.Stop(signalChan)
        log.Debug("Recieved Signal: ", s)
        break LOOP

    case <- spotifyEndChan:
        log.Trace("Recieved on end of track chan")
        updateChan <- Notification{Kind:EndOfTrack}
        log.Trace("Finished send on end of track update")

    case update := <- updateChan:
        log.Trace("Recieved Update: ", update.Kind)

        //Take action based on update type
        switch update.Kind{
        case AddedToQueue:
            track := update.Content.(TrackItem)
            log.Trace("Added Track: "+track.ProviderID)

            //Append
            queue = append(queue, track)

        case RemovedFromQueue:
            //Should have to do nothing...unless is current track
            track := update.Content.(TrackItem)
            log.Warn("Remove Track Not Implemented Yet: "+track.ProviderID)

        case PausedTrack:
            //Send pause command
            log.Trace("Paused Track")

            if(pandoraPlaying){
                pClient.Pause()
            }else{
                sClient.Pause()
            }
            isPlaying = false

        case ResumedTrack:
            //Send play
            log.Trace("Resumed Track")

            if(pandoraPlaying){
                pClient.Play()
            }else{
                sClient.Play()
            }
            isPlaying = true

        case StoppedTrack:
            //Unload current track
            log.Trace("Stopped Track")

            if(pandoraPlaying){
                pClient.Pause()
            }else{
                sClient.Stop()
            }
            isPlaying = false

        case NextTrack:

            // check queue for track, if has one start playing it, otherwise continue
            // pandora
            if len(queue) > 0{
                track := queue[0]

                if pandoraPlaying{
                    pClient.Pause()
                }

                spotifyEndChan = sClient.NextTrack(track)

                queue = queue[1:]
                pandoraPlaying = false

            }else{
                if pandoraPlaying{
                    // Tell Pandora Client to Skip, Handler in Pandora.go will update
                    // other clients with the new song
                    pClient.NextTrack()
                }else{
                    sClient.Stop()

                    if delayedAction != nil {
                        themeId := delayedAction.Content.(string)

                        pClient.PlayStation(themeId)

                        delayedAction := new(Notification)
                    }else{
                        pClient.NextTrack()
                    }

                    pandoraPlaying = true
                }

            }
        case ChangeTheme:
            themeId := update.Content.(string)
            log.Trace("Changed Theme: "+themeId)

            if(pandoraPlaying){
                pClient.PlayStation(themeId)
            }else{
                delayedAction = &update
            }

        case SetVolume:
            volume := update.Content.(uint8)
            log.Trace("Changed Volume: ",volume)

            if pandoraPlaying{
                pClient.SetVolume(volume)
            }else{
                log.Warn("Spotify Volume cannot be set")
            }

        case EndOfTrack:
            // Tell other Clients that the track has ended
            msg := map[string]string{
                "command":"endOfTrack",
            }

            client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,msg)

            updateChan <- Notification{Kind:NextTrack}

        case StatusUpdate:
            //Send back map of current status values: title,isPlaying,queue
            response := map[string]interface{}{
                //"deviceName": deviceName,
                "deviceId": musicBoxID,
                "isPlaying": isPlaying,
                "pandora": pandoraPlaying,
                "queue": queue,
            }

            responseMessage := map[string]interface{}{
                "command": "statusUpdate",
                "data": response,
            }

            client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,responseMessage)
        default:
            log.Warn("Unknown Update Type: %d",update)
        }
    }
}

func initializeClient()(*pandoraClient, *spotifyClient){
    //log.info("Name:" + deviceName + " ID:" + deviceID)

    client = turnpike.NewClient()

    connectToServer(client)
    authenticateClient(client)


    updateChan = make(chan Notification)


    // Subscribe to box url
    client.Subscribe(baseURL+boxUsername+"/"+musicBoxID, handleMessages)


    sClient := SetupSpotify(client)
    pClient := SetupPandora(client)

    return pClient, sClient
}


func handleMessages(topicURI string, event interface{}){
    message := event.(map[string]interface{})
    command := message["command"].(string)

    //fmt.Println("Command: "+command)
    switch command{
    case "addTrack":
        data := message["data"].([]interface{}) // Need for interface due to interal marshaling in turnpike

        //Add all passed tracks
        for _,trackDict := range data {
            track := trackDict.(map[string]interface{})
            newTrack := trackItemFromMap(track)

            updateChan <- Notification{Kind:AddedToQueue,Content:newTrack} // Give chance to preload
        }

    case "playTrack":
        updateChan <- Notification{Kind:ResumedTrack}
    case "pauseTrack":
        updateChan <- Notification{Kind:PausedTrack}
    case "nextTrack":
        updateChan <- Notification{Kind:NextTrack}
    case "updateTheme":
        //Extract station ID
        data := message["data"].(map[string]interface{})
        themeId := data["ThemeID"].(string)

        updateChan <- Notification{Kind:ChangeTheme, Content:themeId}
    case "setVolume":
        // Extract Volume
        data := message["data"].(map[string]interface{})
        volume := uint8(data["Volume"].(float64))

        updateChan <- Notification{Kind:SetVolume, Content: volume}

    case "statusUpdate":
        updateChan <- Notification{Kind:StatusUpdate}
    default:
        log.Warn("Unknown message: ",command)
    }
}




func connectToServer(client *turnpike.Client){
    //Connect socket between server port and local port
    config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
    config.Header.Add("musicbox-box-id", musicBoxID)

    CONNECT:
    if err := client.ConnectConfig(config); err != nil {
        log.Error("Error connecting: ", err)
        time.Sleep(100*time.Millisecond)
        goto CONNECT
    }

    log.Info("Connected to Server at " + serverURL)
}

const WAMP_BASE_URL = "http://api.wamp.ws/"
const WAMP_PROCEDURE_URL = WAMP_BASE_URL+"procedure#"

func authenticateClient(client *turnpike.Client){
    //Start session (lookup user & auth)
    resp := client.Call(baseURL+"musicbox/startSession",musicBoxID)
    message := <-resp


    // Set User Info
    user := message.Result.(map[string]interface{})
    boxUsername = user["username"].(string)
    boxSessionID = user["sessionID"].(string)

    extra := map[string]interface{}{
        "client-type":"musicBox-v1", //Used to diferentiate musicbox from other clients (ie Website)
        "client-id":musicBoxID,
    }

    resp = client.Call(WAMP_PROCEDURE_URL+"authreq",boxUsername,extra)
    message = <-resp

    ch,ok := message.Result.(string)
    if !ok{
        log.Error("Incorrect response type")
    }

    //Calculate & send signature
    sig := authSignature([]byte(ch),boxSessionID,nil)
    resp = client.Call(WAMP_PROCEDURE_URL+"auth",sig)
    <-resp //This give back permissions
}