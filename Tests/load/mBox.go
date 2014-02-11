package main

import(
	"code.google.com/p/go.net/websocket"
	"github.com/cvanderschuere/turnpike"
	"time"
)
type TestBox struct{
	tpClient *turnpike.Client
	authWait chan bool
	boxID string
	username string
	password string	
	sessionID string
	
	isPlaying bool
	
	Recieved chan NotificationType
}


func createBoxClient(boxID string, killChan chan bool)(*TestBox){
	log.Info("Create box: "+boxID)
	
	//
	// Prepare client
	//	
		
	client := TestBox{
		tpClient: turnpike.NewClient(),
		authWait: make(chan bool,1),
		boxID: boxID,
		Recieved: make(chan NotificationType),
	} 
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
	config.Header.Add("musicbox-box-id",boxID)


	CONNECT:
		
	if err := client.tpClient.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		time.Sleep(1)
		goto CONNECT //Loop until connected
	}
		
	//Launch Event handler
	go boxHandler(&client)
	
	//
	// Authenticate
	//
	
	//Start session (lookup user & auth)
	client.tpClient.Call("startSession",baseURL+"musicbox/startSession",boxID)
	
	//Wait until authenticated
	isAuth := <-client.authWait
	if !isAuth{
		log.Error("Failed auth")
		return nil
	}
	
	//
	// Connection authenticated
	//
	
	//Launch pinger to keep websocket open (ELB has 60 second timeout)
	go pingBox(client)
	
	//Subscribe as appropriate
	client.tpClient.Subscribe(baseURL+client.username+"/"+client.boxID)
	
	return &client
}

func boxHandler(client *TestBox){
	
	//Mock track
	/*
	mockTrack := TrackItem{
		Title: "Beautiful Life",
		ArtistName: "Armin Van Buren",
		AlbumName: "Intense",
		ArtworkURL: "www.exampleArtwork.com",
		ProviderID: "spotify:track:3lJbKB0A7wo8HbtlsQep76",
		Length: 1.5,
	}
	*/
	
	for event := range client.tpClient.HandleChan{
		switch event.(type){
		case turnpike.EventMsg:
			message := event.(turnpike.EventMsg).Event.(map[string]interface{})
		
			log.Trace("Box Recieved: "+ message["command"].(string))
			//Switch through command types
			switch message["command"]{
			case "addTrack":
				client.Recieved <- NextTrack
			case "removeTrack":
				client.Recieved <- RemovedFromQueue
			case "playTrack":
				client.isPlaying = true
				client.Recieved <- ResumedTrack
			case "pauseTrack":
				client.isPlaying = false
				client.Recieved <- PausedTrack
			case "stopTrack":
				client.isPlaying = false
				client.Recieved <- StoppedTrack
			case "nextTrack":
				client.Recieved <- NextTrack
			case "updateTheme":
				client.Recieved <- UpdateTheme
			}	
		case turnpike.CallResultMsg:
			message := event.(turnpike.CallResultMsg)
		
			switch message.CallID {
			case "startSession":
				user := message.Result.(map[string]interface{})
				client.username = user["username"].(string)
				client.sessionID = user["sessionID"].(string)
			
				extra := map[string]interface{}{
					"client-type":"musicBox-v1", //Used to diferentiate musicbox from other clients (ie Website)
					"client-id":client.boxID,
				}

				client.tpClient.Call("authreq",WAMP_PROCEDURE_URL+"authreq",client.username,extra)
			case "authreq":
				//Recieve challenge
				ch,ok := message.Result.(string)
				if !ok{
					log.Error("Incorrect response type")
					client.authWait<-false
					continue
				}
			
				//Calculate & send signature
				sig := authSignature([]byte(ch),client.sessionID,nil)
				client.tpClient.Call("auth",WAMP_PROCEDURE_URL+"auth",sig)
			
			case "auth":
				//Recieve permission information
				client.authWait<-true
			}
		default:
			log.Warn("Recieved Unknown type",event)
		}
	}
}

func pingBox(client TestBox){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.tpClient.PublishExcludeMe(baseURL+"ping","blank")
	}
}
