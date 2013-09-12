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
	"github.com/iand/spotify"
	"time"
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
	server.GetAuthSecret = lookupUserSessionID
	server.GetAuthPermissions = getUserPremissions
	server.MessageToPublish = InterceptMessage
	server.OnAuthenticated = userConnected
	
	//Setup RPC Functions (probably not the right way to do this)
	server.RegisterRPC(baseURL+"players",boxRequest)
	server.RegisterRPC(baseURL+"recommendSongs",recommendSongs)
	server.RegisterUnauthRPC(baseURL+"user/startSession",startSession)
	server.RegisterUnauthRPC(baseURL+"musicbox/startSession",startSessionBox)
		
    s := websocket.Server{Handler: postmaster.HandleWebsocket(server), Handshake: VerifyConnection}
	http.Handle("/", s)
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var dynamoDBServer *dynamodb.Server
var usersTable *dynamodb.Table
var musicBoxesTable *dynamodb.Table
var trackHistoryTable *dynamodb.Table

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
	
	//Track History
	primary = dynamodb.NewStringAttribute("CompositeID", "")
	rangeKey := dynamodb.NewStringAttribute("Date", "")
	key = dynamodb.PrimaryKey{primary, rangeKey}
	trackHistoryTable = dynamoDBServer.NewTable("TrackHistory",key)	

	return nil
}


//Verfiy the identify of incoming connection (Accept:nil, Reject: error) sends 403 on error
func VerifyConnection(config *websocket.Config, req *http.Request) (err error){	
	fmt.Println("Verifing connection")
	return nil

	//Get header information to auth connection
	username := req.Header.Get("musicbox-username")
	sessionID := req.Header.Get("musicbox-session-id")
	
	boxID := req.Header.Get("musicbox-box-id")
	
	//Verify if accurate information
	if username != "" && sessionID != "" {
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
	}else if boxID != ""{
		//Sign in as music box
		if box,_ := lookupMusicBox(boxID); box != nil{
			//Find user account info
			if user,_ := lookupUser(box.User); user != nil{
				//Use account information for this user
				config.Header = make(map[string][]string)
				config.Header.Set("musicbox-username",user.Username)
				
				fmt.Println("MusicBoxConnected(%s):%s",boxID,user.Username)
				return nil
			}else{
				return fmt.Errorf("No user associated to this music box")
			}
		}else{
			return fmt.Errorf("invalid music box id")
		}
		
	}
	
	return fmt.Errorf("Invalid identification")
}

//Intercept wamp events (Allow:True, Reject:False)
func InterceptMessage(conn *postmaster.Connection, msg postmaster.PublishMsg)(bool){
	//Filter out base url and split into components
	uri := strings.Replace(msg.TopicURI,baseURL,"",1)
	args := strings.Split(uri,"/")
	
	username := args[0]
	
	data,ok := msg.Event.(map[string]interface{}) //cast
	
	if !ok{
		log.Print("Message doesn't follow correct format: ignoring")
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
	case "startedTrack":
		log.Print("Track Started")
		//Parse recieved track
		d := data["data"].(map[string]interface{})
		
		track := d["track"].(map[string]interface{})
		t := TrackItem{ProviderID:track["ProviderID"].(string),Title:track["Title"].(string),ArtistName:track["ArtistName"].(string),AlbumName:track["AlbumName"].(string),ArtworkURL:track["ArtworkURL"].(string)}
		
		deviceID := d["deviceID"].(string)
		
		//Create aws item
		atts := []dynamodb.Attribute{
			*dynamodb.NewStringAttribute("Title",t.Title),
			*dynamodb.NewStringAttribute("ArtistName",t.ArtistName),
			//*dynamodb.NewStringAttribute("AlbumName",t.AlbumName), //Moment.us doesn't always provide this
			//*dynamodb.NewStringAttribute("ArtworkURL",t.ArtworkURL), //Moment.us doesn't always provide this
			*dynamodb.NewStringAttribute("ProviderID",t.ProviderID),
		}
								
		//Add track to database for this user:musicbox
		if _,err := trackHistoryTable.PutItem(username+":"+deviceID,time.Now().UTC().Format(time.RFC822Z),atts); err != nil{
			log.Print(err.Error())
		}else{
			log.Print("Put New track")
		}
		
		
	default:
			log.Print("Unknown Command:",data["command"])
	}
		
	fmt.Println(username,data)	
		
	return true
}

//
// RPC Handlers
//


//Return music box device names for given user (need auth down the line)
func boxRequest(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){	
	
	var boxes []BoxItem
	
	//Look up music box ids for user
	if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: conn.Username}); err == nil{
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

// Args format [boxid]
func recommendSongs(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){	
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
	
	user,err2 := lookupUser(conn.Username)
	if err2 != nil{
		return nil,err2
	}else if box.User != user.Username{
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid boxID",Details:""}
	}
	
	//Make Moment.us request based on box information
	fmt.Println(box)
	
	v := url.Values{}
	v.Set("access_token", user.SessionID)
	
	//2006-01-02T15:04:05Z time format layout time.RFC3339
	v.Set("current_context[date]",time.Now().Format(time.RFC3339))
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
	
	var tracks []*TrackItem
	c := spotify.New() //Spotify client
	
	Recommend_Loop:
	for _,track := range responseObject.Data{
		//Turn Momemtus track into TrackItem
		t := &TrackItem{}
		t.Title = track.Title
		t.ArtistName = track.Artist.Name
		t.AlbumName = track.Album.Name
		if len(track.Artist.Image) > 0{
			t.ArtworkURL = track.Artist.Image[0].Content
		}
		
		//Look up on spotify
		if r,e := c.SearchTracks(t.Title+" "+t.ArtistName,0); e == nil{
			
			//Make sure results returned
			if r.Info.TotalResults > 0{
				Result_Loop:
				for _,track := range r.Tracks{
					if strings.Contains(track.Album.Availability.Territories,"US"){
						t.ProviderID = track.URI
						break Result_Loop
					}
				}
			}else{
				fmt.Println("Spotify error: no matching track")
				continue Recommend_Loop
			}
			
		}else{
			fmt.Println("Spotify error: ",e)
			continue Recommend_Loop
		}
		
		tracks = append(tracks,t)
	}
	
	return tracks,nil
}

// Used as a login
func startSession(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	searchUser := args[0].(map[string]interface{})
	
	user,err := lookupUser(searchUser["username"].(string))
	if err != nil{
		return nil,err
	}
	
	//Check password
	if user.Password == searchUser["password"].(string){
		//send back session id
		res := map[string]string{
			"username":user.Username,
			"sessionID":user.SessionID,
		}
		
		return res,nil
		
	}else{
		return nil,nil
	}
	
}
func startSessionBox(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	//lookup musicbox
	box,err := lookupMusicBox(args[0].(string))	
	if err != nil{
		//Do something
		return nil,nil
	}
	
	//Lookup user
	user,err := lookupUser(box.User)
	if err != nil{
		return nil,err
	}
	
	res := map[string]string{
		"username":user.Username,
		"sessionID":user.SessionID,
	}
	
	return res,nil
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

//
// Auth
//

func lookupUserSessionID(authKey string)(string,error){
	fmt.Println("Lookup user: "+authKey);
	
	user,err := lookupUser(authKey)
	if err == nil{
		return user.SessionID,nil
	}else{
		return "",errors.New("Can't find user")
	}
}

func getUserPremissions(authKey string,authExtra map[string]interface{})(postmaster.Permissions,error){
	
	p := postmaster.Permissions{
		RPC:map[string]postmaster.RPCPermission{
			baseURL+"currentQueueRequest":true,
			baseURL+"players":true,
			baseURL+"recommendSongs":true,
		},
		PubSub:map[string]postmaster.PubSubPermission{
		},
	}
	user,err := lookupUser(authKey)
	if err == nil{
		//Add pubSub for all music boxes [base+username+boxid]
		for _,boxID := range user.MusicBoxes{
			p.PubSub[baseURL+authKey+"/"+boxID] = postmaster.PubSubPermission{true,true}
		}
	}
	
	return p,nil
}

func userConnected(authKey string, permission postmaster.Permissions){
	fmt.Println("Connected user: "+authKey)
	
	
}


