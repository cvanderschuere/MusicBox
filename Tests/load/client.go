package main

import(
	"code.google.com/p/go.net/websocket"
	"github.com/cvanderschuere/turnpike"
	"time"
	"fmt"
)

type SendInfo struct{
	noti NotificationType
	id	string
}

type TestClient struct{
	tpClient *turnpike.Client
	authWait chan bool
	username string
	password string	
	sessionID string
	permissions []string
	
	//Channels
	Send chan SendInfo
	Recieved chan NotificationType
}


func createClient(username,password string, killChan chan bool)(*TestClient){	
	fmt.Println("Created Client for: ",username)
	
	//
	// Prepare client
	//	
	
	client := TestClient{
		tpClient: turnpike.NewClient(),
		authWait: make(chan bool,1),
		username: username,
		password: password,
		Send: make(chan SendInfo),
		Recieved: make(chan NotificationType),
	}
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")

	CONNECT:
		
	if err := client.tpClient.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		time.Sleep(1)
		goto CONNECT //Loop until connected
	}
		
	//Launch Event handler
	go clientEvents(&client)
	
	
	//
	// Authenticate
	//
	
	m := map[string]string{
		"username" : username,
		"password" : password,
	}
	
	//Start session (lookup user & auth)
	client.tpClient.Call("startSession",baseURL+"user/startSession",m)
	
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
	go pingTestClient(client)
	
	
	//Subscribe as appropriate	
	for _,value := range client.permissions{
		client.tpClient.Subscribe(value)
	}
	
	//Lauch send chans
	go sendNotificationsFromClient(client)
	
	return &client
}

func clientEvents(client *TestClient){
	
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
		
			log.Trace("Client Recieved: "+ message["command"].(string))
			//Switch through command types
			switch message["command"]{
			case "addTrack":
				client.Recieved <- NextTrack
			case "removeTrack":
				client.Recieved <- RemovedFromQueue
			case "playTrack":
				client.Recieved <- ResumedTrack
			case "pauseTrack":
				client.Recieved <- PausedTrack				
			case "stopTrack":
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
				client.sessionID = user["sessionID"].(string)
			

				client.tpClient.Call("authreq",WAMP_PROCEDURE_URL+"authreq",client.username,map[string]interface{}{})
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
				
				p := message.Result.(map[string]interface{})
				rpc := p["PubSub"].(map[string]interface{})
				
				client.permissions = make([]string, len(rpc))
				
				i := 0
				for key,_ := range rpc{
					client.permissions[i] = key
					i++
				}
				
				//Recieve permission information
				client.authWait<-true
			}
		case turnpike.CallErrorMsg:
			message := event.(turnpike.CallErrorMsg)
		
			switch message.CallID {
			case "startSession":
				client.authWait<-false
			}
			
		default:
			log.Warn("Recieved Unknown type",event)
		}
	}
}

func sendNotificationsFromClient(client TestClient){
	for info := range client.Send{
	switch info.noti{
		case NextTrack:
		   msg := map[string]interface{} {
	               "command":"nextTrack",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)
		   
		case RemovedFromQueue:
		   msg := map[string]interface{} {
	               "command":"removeTrack",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)
		case ResumedTrack:
		   msg := map[string]interface{} {
	               "command":"playTrack",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)
		case PausedTrack:
		   msg := map[string]interface{} {
	               "command":"pauseTrack",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)
		case StoppedTrack:
		   msg := map[string]interface{} {
	               "command":"stopTrack",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)
		case UpdateTheme:
		   msg := map[string]interface{} {
	               "command":"updateTheme",
	       }
		   client.tpClient.PublishExcludeMe(baseURL+client.username+"/"+info.id,msg)	
		}	
	}
}

func pingTestClient(client TestClient){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.tpClient.PublishExcludeMe(baseURL+"ping","blank")
	}
}
