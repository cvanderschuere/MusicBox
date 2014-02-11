package main

import (
	"github.com/jcelliott/lumber"
	"strconv"
	"sync"
	"time"
	"fmt"
)

const serverURL = "ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"
const baseURL = "http://www.musicbox.com/"

const testUsername = "testuser"
const testPassword = "testPassword"


//Auth info
const WAMP_BASE_URL = "http://api.wamp.ws/"
const WAMP_PROCEDURE_URL = WAMP_BASE_URL+"procedure#"
var authWait = make(chan bool,1) //Used to block until authentication

var log = lumber.NewConsoleLogger(lumber.INFO)

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
	UpdateTheme
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

/*
	Functions
*/

func main() {	
	
	testAuth()
		
	//Run each test
	num := 1
	iter := 1
	
	latency := testLoopbackTime(num,iter);
	fmt.Printf("Total time: %fs avg for %d clients/devices with %d loops\n",latency,num,iter)
	fmt.Printf("Averge per loop: %fs \n",latency/float64(iter))

}

/*
	Standard notification looop

	for val := range updateChan
		log.Trace("Update: ",update.Kind)
		
		//Take action based on update type
		switch update.Kind{
		case AddedToQueue:
		case RemovedFromQueue:
		case PausedTrack:
			
		case ResumedTrack:
		case StoppedTrack:
			
		case NextTrack:
		default:
			log.Warn("Unknown Update Type: %d",update)
		}	
	}
*/

func testAuth()(error){
	
	fmt.Println("Incorrect Login")
	createClient("SampleUser","IncorrectPassword", nil)
	
	fmt.Println("Correct Login")
	createClient(testUsername,testPassword, nil)
	
	return nil
}


func testLoopbackTime(num int, loopCount int)(float64){
	
	mBoxes := make([]*TestBox,num)
	clients := make([]*TestClient,num)
	
	//Scale number of concurrent clients each loop
	for i := 0; i<num; i++ {
		//
		//Create music box and client
		//
		
		//Mbox
		mBoxes[i] = createBoxClient("testBoxID"+strconv.Itoa(i+1),nil)
	
		//Matching Client
		clients[i] = createClient(testUsername,testPassword, nil)
		
	}
	
	//
	// Send/recieve play/pause command in loop
	//
	
	deltas := make([]float64,num)
	wg := new(sync.WaitGroup)
	
	
	for i,_ := range mBoxes{
		wg.Add(1)
		go internalLatencyTest(i,loopCount, clients[i], mBoxes[i],deltas,wg)
	}	
		
		
	//Wait for waitgroup to finish
	wg.Wait()
	
	
	//Sum & average deltas
	sum := 0.0
	for _,delta := range deltas{
		sum += delta
	}
	
	sum /= float64(len(deltas)); // sum = sum/len(deltas)
	
	
	//Return average	
	return sum
}

func  internalLatencyTest(i,loopCount int, client *TestClient, box *TestBox, deltas []float64, wg *sync.WaitGroup){					
		startTime := time.Now()
		
		for a := 0; a<loopCount; a++{
			//Send play to box
			client.Send <- SendInfo{
								noti:NextTrack,
								id: box.boxID,
							}
		
			//wait for box to recieve play
			<-box.Recieved
		}
		
		//Store delay in deltas
		deltas[i] = time.Since(startTime).Seconds()
		
		wg.Done()
}
	



