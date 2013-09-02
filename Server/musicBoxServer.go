package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"fmt"
	"net/http"
	"errors"
	//"github.com/cvanderschuere/turnpike"
	"postmaster"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"strings"
)

//Global
var server *postmaster.Server

const(
	baseURL = "http://www.musicbox.com/"
)

func main() {
	
	if err := setupAWS();err != nil{
		log.Fatal("AWS Login Error: err")
		return
	}
	
	server = postmaster.NewServer()

	//Customize server functionality
	server.MessageToPublish = InterceptMessage
	
	//Setup RPC Functions (probably not the right way to do this)
	server.RegisterRPC(baseURL+"currentQueueRequest",queueRequest)
	server.RegisterRPC(baseURL+"players",boxRequest)
	//	server.RegisterRPC(baseURL+"user/status",userUpdate)
	//	server.RegisterRPC(baseURL+"player/status",playerUpdate)
	
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: VerifyConnection}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var dynamoDBServer *dynamodb.Server

func setupAWS()(error){
	//Sign in to AWS
	auth, err := aws.EnvAuth()

	if err != nil {
		return errors.New("Signin to AWS failed: "+err.Error())
	}

	dynamoDBServer = &dynamodb.Server{auth, aws.USWest2}
		
	//Users
	primary := dynamodb.NewStringAttribute("username", "")
	key := dynamodb.PrimaryKey{primary, nil}
	usersTable = dynamoDBServer.NewTable("Users",key)	

	return nil
}

var usersTable *dynamodb.Table

//Verfiy the identify of incoming connection (Accept:nil, Reject: error) sends 403 on error
func VerifyConnection(config *websocket.Config, req *http.Request) (err error){	

	//Get header information to auth connection
	username := req.Header.Get("musicbox-username")
	sessionID := req.Header.Get("musicbox-session-id")
	
	//Verify if accurate information
	if username != "" && sessionID != ""{
		//Get user information from database
		if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: username}); err == nil{
			userObj := &UserItem{}

			err := dynamodb.UnmarshalAttributes(&item, userObj)
			if err != nil {
				return fmt.Errorf("Error with object unmarshelling")
			}
			
			//Compare session id to verify accuracy
			if userObj.SessionID != sessionID{
				return fmt.Errorf("Invalid sessionID")
			}
			
			//Need to pass username information on to handler (Somewhat of a hack; better options welcome)
			config.Header = make(map[string][]string)
			config.Header.Set("musicbox-username",username)
						
			return nil	//Verified connection	
		}else{
			return fmt.Errorf("Invalid username")
		}
	}
	
	return fmt.Errorf("Invalid identification")
}

//Intercept wamp events (Allow:True, Reject:False)
func InterceptMessage(id postmaster.ConnectionID, msg postmaster.PublishMsg)(bool){
	//Filter out base url and split into components
	uri := strings.Replace(msg.TopicURI,baseURL,"",1)
	args := strings.Split(uri,"/")
	
	username := args[0]
	
	data,ok := msg.Event.(map[string]interface{}) //cast
	
	if !ok{
		log.Print("Doesn't follow correct formate")
		return false
	}
	
	//Switch through command types
	switch data["command"]{
	case "addTrack":
	case "removeTrack":
	case "playTrack":
	case "pauseTrack":
	case "stopTrack":
	case "nextTrack":
	default:
			log.Print("Unknown Command:",data["command"])
	}
		
	fmt.Println(username,data)	
		
	return true
}

//
// RPC Han
//

//RPC Handler of form: res, err = f(id, msg.ProcURI, msg.CallArgs...)
func queueRequest(id postmaster.ConnectionID,username string, url string, args ...interface{})(interface{},*postmaster.RPCError){
	//Format: [deviceName]
	deviceName := args[0].(string)
	
	//Recieved request for queue...for now just pass on to music box
	
	//This will forward an event on a private channel to the music box
	//The music box will then publish a typical CurrentQueue update to everyone
	statusMsg := map[string]string{
		"command":"statusUpdate",
	}
	
	server.SendEvent(baseURL+username+"/"+deviceName+"/internal",statusMsg);
	
	//No response necessary
	return nil,nil
}

//Return music box device names for given user (need auth down the line)
func boxRequest(id postmaster.ConnectionID,username string,url string, args ...interface{})(interface{},*postmaster.RPCError){	
	
	//Simulate for  now
	players := []string{"Awolnation","Beatles","Coldplay","Deadmau5"}
	
	return players,nil
}

