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

	tables, err := dynamoDBServer.ListTables()

	if err != nil {
		return err
	}
	
	log.Print(tables)
	
	//Create tables
	primary := dynamodb.NewStringAttribute("username", "")
	key := dynamodb.PrimaryKey{primary, nil}
	usersTable = dynamoDBServer.NewTable(tables[0],key)
	
	userTableStats,err := usersTable.DescribeTable()
	if err != nil{
		log.Fatal(err)
	}else{
		log.Printf("%s(%s): %d",userTableStats.TableName,userTableStats.TableStatus,userTableStats.ItemCount)
	}

	return nil
}

var usersTable *dynamodb.Table

//Verfiy the identify of incoming connection (Accept:true, Reject: false)
func VerifyConnection(config *websocket.Config, req *http.Request) (err error){	

	username := req.Header.Get("musicbox-username")
	sessionID := req.Header.Get("musicbox-session-id")
	if username != "" && sessionID != ""{
		log.Printf("Username: %s SessionID: %s\n",username,sessionID)
		return nil
	}else{
		return fmt.Errorf("Invalid identification")
	}	
}

//Intercept wamp events (Allow:True, Reject:False)
func InterceptMessage(id postmaster.ConnectionID, msg postmaster.PublishMsg)(bool){
	return true
}


//RPC Handler of form: res, err = f(id, msg.ProcURI, msg.CallArgs...)
func queueRequest(id postmaster.ConnectionID, url string, args ...interface{})(interface{},*postmaster.RPCError){
	//Format: [username password(hashed) deviceName]
	username := args[0].(string)
	//password := args[1].(string)
	deviceName := args[2].(string)
	
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
func boxRequest(id postmaster.ConnectionID,url string, args ...interface{})(interface{},*postmaster.RPCError){
	//Format: [username password(hashed)]
	//username := args[0].(string)
	//password := args[1].(string)
	
	//Simulate for  now
	players := []string{"Awolnation","Beatles","Coldplay","Deadmau5"}
	
	return players,nil
}

