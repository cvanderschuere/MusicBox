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
	"sort"
	"net/url"
	"io/ioutil"
	"encoding/json"
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
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	//	server.RegisterRPC(baseURL+"user/status",userUpdate)
	//	server.RegisterRPC(baseURL+"player/status",playerUpdate)
	
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: VerifyConnection}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var dynamoDBServer *dynamodb.Server
var usersTable *dynamodb.Table
var musicBoxesTable *dynamodb.Table

func setupAWS()(error){
	//Sign in to AWS
	auth, err := aws.EnvAuth()

	if err != nil {
		return errors.New("Signin to AWS failed: "+err.Error())
	}

	dynamoDBServer = &dynamodb.Server{auth, aws.USWest2}
		
	//Users
	primary := dynamodb.NewStringAttribute("Username", "")
	key := dynamodb.PrimaryKey{primary, nil}
	usersTable = dynamoDBServer.NewTable("Users",key)	
	
	//MusicBoxes
	primary = dynamodb.NewStringAttribute("ID", "")
	key = dynamodb.PrimaryKey{primary, nil}
	musicBoxesTable = dynamoDBServer.NewTable("MusicBoxes",key)	

	return nil
}


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
	
	var boxes []BoxItem
	
	//Look up music box ids for user
	if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: username}); err == nil{
		userObj := &UserItem{}

		err := dynamodb.UnmarshalAttributes(&item, userObj)
		if err != nil {
			err2 := &postmaster.RPCError{URI:url,Description:"Unmarshal Error",Details:""}
			return nil,err2
		}
						
		//Batch lookup ids for music boxes
		for _,id := range userObj.MusicBoxes{
			boxObj := &BoxItem{}
			
			//Get music box
			if box, err3 := musicBoxesTable.GetItem(&dynamodb.Key{HashKey: id}); err3 == nil{
				boxErr := dynamodb.UnmarshalAttributes(&box, boxObj)
				if boxErr != nil {
					boxErr2 := &postmaster.RPCError{URI:url,Description:"Unmarshal Error",Details:""}
					return nil,boxErr2
				}
			}
			
			boxes = append(boxes,*boxObj)
		}
		
	}else{
		err2 := &postmaster.RPCError{URI:url,Description:"Get error:invalid user",Details:""}
		return nil, err2
	}
	
	// FIXME Limit response to match old api
	var players sort.StringSlice;
	for _,box := range boxes{
		players = append(players,box.ID)
	}
	
	//Sort
	players.Sort() 
		
	return players,nil
}


func recommendSongs(id postmaster.ConnectionID,username string,uri string, args ...interface{})(interface{},*postmaster.RPCError){	
	//Look up music box with ID
	boxID,ok := args[0].(string)
	if !ok{
		//Incorrect format
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format",Details:""}
	}
	
	box,err := lookupMusicBox(boxID)
	if err != nil{
		return nil,err
	}
	
	user,err2 := lookupUser(username)
	if err2 != nil{
		return nil,err2
	}else if box.User != user.Username{
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid boxID",Details:""}
	}
	
	//Make Moment.us request based on box information
	fmt.Println(box)
	
	v := url.Values{}
	v.Set("access_token", user.SessionID)
	v.Set("current_context[date]","2013-09-03T03:09:31Z")
	v.Set("current_context[location][lng]",box.Location[0])
	v.Set("current_context[location][lat]",box.Location[1])
	query := v.Encode()

	resp,errGet := http.Get("https://api.wearemoment.us/v1/songs/discover?"+query)
	if errGet != nil {
		// handle error
		fmt.Println("Moment.us Error:%s",err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
		
	//Unmarshal JSON
	var responseObject DiscoverResult
	errJson := json.Unmarshal(body,&responseObject)
	if errJson != nil {
			fmt.Println("Json Error: %s",err)
			fmt.Println(string(body))
			return nil, nil
	}
	
	for _,track := range responseObject.Data{
		fmt.Println(track)
	}
	
	return nil,nil
}


func lookupMusicBox(id string)(*BoxItem,*postmaster.RPCError){
	boxObj := &BoxItem{}
	
	//Get music box
	if box, err3 := musicBoxesTable.GetItem(&dynamodb.Key{HashKey: id}); err3 == nil{
		boxErr := dynamodb.UnmarshalAttributes(&box, boxObj)
		if boxErr != nil {
			boxErr2 := &postmaster.RPCError{URI:"",Description:"Unmarshal Error",Details:""}
			return nil,boxErr2
		}else{
			return boxObj,nil
		}
	}else{
		err2 := &postmaster.RPCError{URI:"",Description:"Invalid BoxID",Details:""}
		return nil, err2
	}
}

func lookupUser(username string)(*UserItem,*postmaster.RPCError){
	if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: username}); err == nil{
		userObj := &UserItem{}

		err := dynamodb.UnmarshalAttributes(&item, userObj)
		if err != nil {
			err2 := &postmaster.RPCError{URI:"",Description:"Unmarshal Error",Details:""}
			return nil,err2
		}else{
			return userObj,nil
		}
	}else{
		err2 := &postmaster.RPCError{URI:"",Description:"Get error:invalid user",Details:""}
		return nil, err2
	}
}

