package main

import(
	"time"
	"strings"
	"postmaster"
	"log"
	"github.com/crowdmob/goamz/dynamodb"
	"fmt"
)

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