package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	//"github.com/cvanderschuere/turnpike"
	"postmaster"
	//"turnpike" //Local Dev
)

//Global
var server *postmaster.Server

const(
	baseURL = "http://www.musicbox.com/"
)

func main() {
	server = postmaster.NewServer()
	
	//Setup RPC Functions (probably not the right way to do this)
	server.RegisterRPC(baseURL+"currentQueueRequest",queueRequest)
	server.RegisterRPC(baseURL+"players",boxRequest)
	//	server.RegisterRPC(baseURL+"user/status",userUpdate)
	//	server.RegisterRPC(baseURL+"player/status",playerUpdate)
	
	http.Handle("/", websocket.Handler(postmaster.HandleWebsocket(server)))
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
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

