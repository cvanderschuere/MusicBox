package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/cvanderschuere/turnpike"
	"io/ioutil"
	"log"
	"net/http"
	//"turnpike" //Local Dev
)

//Global
var server *turnpike.Server

const (
	baseURL = "http://www.musicbox.com/"
)

func main() {
	server = turnpike.NewServer()

	//Setup RPC Functions (probably not the right way to do this)
	server.RegisterRPC(baseURL+"currentQueueRequest", queueRequest)
	server.RegisterRPC(baseURL+"players", boxRequest)
	//	server.RegisterRPC(baseURL+"user/status",userUpdate)
	//	server.RegisterRPC(baseURL+"player/status",playerUpdate)

	http.Handle("/", websocket.Handler(turnpike.HandleWebsocket(server)))

	go startWebServer()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func startWebServer() {
	webServer := http.NewServeMux()

	webServer.HandleFunc("/", serveHomePage)

	if err := http.ListenAndServe(":80", webServer); err != nil {
		log.Fatal("Unable to Start Web Server: ", err)
	}
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile("/home/ubuntu/MusicBoxWebClient/index.html")

	if err != nil {
		return
	}

	w.Write([]byte(body))
}

//RPC Handler of form: res, err = f(id, msg.ProcURI, msg.CallArgs...)
func queueRequest(id, url string, args ...interface{}) (interface{}, error) {
	//Format: [username password(hashed) deviceName]
	username := args[0].(string)
	//password := args[1].(string)
	deviceName := args[2].(string)

	//Recieved request for queue...for now just pass on to music box

	//This will forward an event on a private channel to the music box
	//The music box will then publish a typical CurrentQueue update to everyone
	statusMsg := map[string]string{
		"command": "statusUpdate",
	}

	server.SendEvent(baseURL+username+"/"+deviceName+"/internal", statusMsg)

	//No response necessary
	return nil, nil
}

//Return music box device names for given user (need auth down the line)
func boxRequest(id, url string, args ...interface{}) (interface{}, error) {
	//Format: [username password(hashed)]
	//username := args[0].(string)
	//password := args[1].(string)

	//Simulate for  now
	players := []string{"Awolnation", "Beatles", "Coldplay", "Deadmau5"}

	return players, nil
}
