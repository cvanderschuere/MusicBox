package speakeasy

import (
	"github.com/cvanderschuere/alsa-go"
	"avahi-go"
	"log"
	"code.google.com/p/go.net/websocket"
	"net/http"
	"strconv"
	"time"
	"sync"
	"errors"
	"fmt"
)

//
// Types
//

//Avahi
const(
	serviceType = "_musicbox._tcp"
	port = 8060
)
//Message
type MessageType uint8
const(
	_ MessageType = iota
	ERROR
	TIME
	NEW_STREAM
	DATA
)

const WS_BACKLOG = 2 //Num messages to be queued

type Speaker struct{
	//Name
	username string
	name	string
	
	//Input Stream
	Input chan alsa.AudioStream
	Control chan bool
		
	//OutputStream
	outputStreams [](chan []byte)
	outputControls []chan bool
	
	//Alsa (internal)
	controlChan	chan bool
	streamChan	chan alsa.AudioStream
	
	//Sync
	hardwareDelay time.Duration//Recieved from alsa	
	peerLock	*sync.Mutex
	
	//External Connections
	externalSpeakers	map[string]*externalSpeaker //Maps speaker name to speaker instance
	killPublishChan		chan interface{}
	killBrowseChan		chan interface{}
	
	//Recieve message
	ExternalSpeakerChan	chan Message
	
}

type externalSpeaker struct{
	username string
	name	 string
	
	//Sync
	networkDelay time.Duration //Calculated from TIME messages
	
	//Published services
	service avahi.Service
	
	//Websocket
	DataChan chan interface{}
	ws	*websocket.Conn
}
type handshake struct{
	username string
	name string
}

type Message struct{
	Kind	MessageType
	time	time.Time
	Data	interface{}
}

//Global
var currentSpeaker *Speaker

func NewSpeaker(username,name string)(*Speaker,error){
	var err error
	
	//Remove existing speaker if exists
	if currentSpeaker != nil{
		currentSpeaker.Cleanup()
	}
	
	//Create new speaker
	currentSpeaker = new(Speaker)
	
	//Create input
	currentSpeaker.Input = make(chan alsa.AudioStream)
	
	//Configure output
	currentSpeaker.controlChan = make(chan bool)
	currentSpeaker.streamChan = alsa.Init(controlChan)
	
	//Configure multiplexing
	currentSpeaker.outputStreams = make([](chan []byte),1)
	currentSpeaker.outputControls = make([]chan bool,1)
	
	//Init internal elements
	currentSpeaker.username = username
	currentSpeaker.name = name
	currentSpeaker.externalSpeakers = make(map[string]externalSpeaker)
	currentSpeaker.recieveChan = make(chan Message,WS_BACKLOG)
	currentSpeaker.peerLock = new(sync.Mutex)
	
	//Publish speaker with this name on network
	currentSpeaker.killPublishChan,err = avahi.PublishService(name,serviceType,port)
	if err != nil{
		log.Fatal(err)
	}
	
	//Avahi browse for peers
	currentSpeaker.killBrowseChan = make(chan interface{})
	go startBrowse(serviceType)
	
	
	
	
	return currentSpeaker,nil
}


//
//	Object methdos
//

//Configure of multi output: default on device output only
func (speaker *Speaker) ConfigureOutputSpeakers(speakerNames []string)(error){
	
	
	
	for _,value := range speakerNames{
		//Match speaker names to speakers 
		
		
	}	
	
	
	//Change values of outputStreams accordingly
	
}


//Closed and cleans up after itself
func (speaker *Speaker) Cleanup(){
	
	//Kill avahi 
	speaker.killPublishChan <- true
	speaker.killBrowseChan <- true
	
	//Close input stream
	close(speaker.Input)
	
}

//
// Internal
//
	
//Search for other peers
func startBrowsing(serviceType string){	
	//Wait while others try and connect to it
	time.Sleep(50 * time.Millisecond)
	
	resultChan := avahi.BrowseService(serviceType,currentSpeaker.killBrowseChan)
	
	//Loop through all updates
	for result := range resultChan{
		//Lock access to external...prevent clash between adding each other
		currentSpeaker.peerLock.Lock()
		
		for key,value := range result{
			//Check if speaker exists
			extern,ok := currentSpeaker.externalSpeakers[key]
			if ok == false{
				//Make new external speaker
				extern = new(externalSpeaker)
				
				//Save
				currentSpeaker.externalSpeakers[key] = extern
			}
			
			//Match service up to entry
			extern.service = value
		}
		
		//Unlock access becausef updates could be far apart
		currentSpeaker.peerLock.Unlock()
	}	
}

	
//Listen to incoming peer connections	
func startListener(port int){
	//Listen to websockets on port
	http.Handle("/", websocket.Handler(handleWebsocket))
	portString,_ := strconv.Itoa(port) 
	
	err := http.ListenAndServe(":"+portString, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

//Handle websocket request
func handleWebsocket(ws *websocket.Conn){
	defer ws.Close()
	
	//Verify 
	var data handshake
	websocket.JSON.Recieve(ws,&data)
	
	if currentSpeaker.username != data.username{
		//Cannot connect to this player...return error
		errResponse := Message{Kind:ERROR,Data:"Incorrect username"}
		websocket.JSON.Send(ws,errResponse)
		return
	}
	
	//Make channels to send/recieve from this speaker
	dataChan :=  make(chan []byte,WS_BACKLOG)
	defer close(dataChan)
	
	//Concurrently send messages
	go func(){
		//Send all messages 
		for data := range dataChan{
			//Get current time
			currentTime := time.Now()
			
			switch data{
			//If  nil...just send time
			case nil:
				msg := Message{Kind:TIME,time:currentTime}
				websocket.JSON.Send(ws,msg)
			default:
				//Send time and data
				msg := Message{Kind:DATA,time:currentTime,Data:data}
				websocket.JSON.Send(ws,msg)
			}
		}
	}()
	
	//Check if speaker exists
	currentSpeaker.peerLock.Lock()
	speaker,ok := currentSpeaker.externalSpeakers[data.name]
	if ok != true{
		//Create new speaker
		speaker = &externalSpeaker{username:data.username,name:data.name}
		
		//Save
		currentSpeaker.externalSpeakers[data.name] = speaker
	}
	else{
		//Close previous connection
		close(speaker.DataChan)
		speaker.ws.Close()
		
	}
	
	//Modify for new connection
	speaker.DataChan = dataChan
	speaker.ws = ws
	
	currentSpeaker.peerLock.Unlock()
	
	
	//Send initial time sync
	dataChan<-nil
	
	//Periodically send time updates
	go func(){
		c := time.Tick(1 * time.Minute)
		for now := range c {
		    //Send time update
			dataChan<-nil
		}
	}
	
	//Recieve on websocket
	var msg Message
	for{
		err := websocket.JSON.Receive(conn, &msg)
		if err != nil {
			log.Error("Error receiving message, aborting connection: %s", err)
			return
		}
		
		//Pass message upstream
		currentSpeaker.ExternalSpeakerChan<-msg
	}
}

/*
func recieveOnConn(conn *websocket.Conn, speaker *externalSpeaker){
	var msg message
	//Loop for life of socket
	for{
		err := websocket.JSON.Receive(conn, &msg)
		if err != nil {
			log.Error("Error receiving message, aborting connection: %s", err)
			return
		}
		
		switch msg.kind{
		case ERROR:
			fmt.Println("Error from external(%s):",speaker.name + msg.data)
		case TIME:
			//Update time delay
			speaker.networkDelay = time.Since(msg.time)
		case NEW_STREAM:
			//Must be in idle status
			if currentSpeaker.Status != IDLE{
				//Wait for a period of time and try again
				<-time.After(2*time.Second)
				
				if currentSpeaker.Status != IDLE{
					//Otherwise send error
					errResponse := Message{Kind:ERROR,Data:"Player not idle; must stop player before proceeding"}
					websocket.JSON.Send(ws,errResponse)
				}
			}
			
			//Configure for new stream
			
			
		case DATA:
			//Feed data into current stream
		}		
	}
}
*/
