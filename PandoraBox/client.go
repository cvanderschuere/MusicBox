package main

import(
	"code.google.com/p/go.net/websocket"
	"github.com/jcelliott/lumber"
	"github.com/cvanderschuere/turnpike"
	"os"
	"os/signal"
	"time"
	"fmt"
)

//
// Constants
//

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"
const musicBoxID = "musicBoxID1"

//Auth info
const WAMP_BASE_URL = "http://api.wamp.ws/"
const WAMP_PROCEDURE_URL = WAMP_BASE_URL+"procedure#"
var boxUsername string
var boxSessionID string
var client *turnpike.Client
var log = lumber.NewConsoleLogger(lumber.TRACE)
var deviceName,_ = os.Hostname()


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

func main() {
		
	setupClient()	
		
	//Register for signals
	signalChan := make(chan os.Signal,1)
	signal.Notify(signalChan)
		
	//Wait on signal
	s := <-signalChan
	signal.Stop(signalChan)
	log.Debug("Recieved Signal: ", s)	
}

func setupClient()(*turnpike.Client){
	log.Info("Name: "+deviceName+"ID: "+musicBoxID)
		
	//
	// Prepare client
	//	
		
	client = turnpike.NewClient()
	
	//Connect socket between server port and local port
	config,_ := websocket.NewConfig("ws://"+serverURL,"http://localhost:4040")
	config.Header.Add("musicbox-box-id",musicBoxID)

	CONNECT:
	if err := client.ConnectConfig(config); err != nil {
		log.Error("Error connecting: ", err)
		time.Sleep(100*time.Millisecond)
		goto CONNECT
	}
	
	fmt.Println("CONNECTED TO SERVER")
	
	//
	// Authenticate
	//
	authenticate(client)

	//
	// Connection authenticated
	//
	
	//Launch pinger to keep websocket open (ELB has 60 second timeout)
	go pingClient(client)
	
	//Subscribe as appropriate
	fmt.Println(baseURL+boxUsername+"/"+musicBoxID)
	client.Subscribe(baseURL+boxUsername+"/"+musicBoxID, handleEvent)
	
	setupPandora(client)
	
	return client
}

func authenticate(client *turnpike.Client){
	//Start session (lookup user & auth)
	resp := client.Call(baseURL+"musicbox/startSession",musicBoxID)
	message := <-resp

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

func pingClient(client *turnpike.Client){
	t := time.Tick(50 * time.Second)
	
	for _ = range t{
		client.PublishExcludeMe(baseURL+"ping","blank")
	}
}
