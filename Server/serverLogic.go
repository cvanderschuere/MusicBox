package main

import(
	"time"
	"strings"
	"postmaster"
	"log"
	"github.com/crowdmob/goamz/dynamodb"
	"strconv"
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
		//Parse recieved tracks
		tracks := data["data"].([]interface{})

		for _,m := range tracks{
			track := m.(map[string]interface{})

			t := TrackItem{
					ProviderID:track["ProviderID"].(string),
					Title:track["Title"].(string),
					ArtistName:track["ArtistName"].(string),
					AlbumName:track["AlbumName"].(string),
					ArtworkURL:track["ArtworkURL"].(string),
					Length:track["Length"].(float64),
			}

			addTrackToCachedQueue(args[1], t)
			addTrackToQueue(args[1],t)
		}

	case "removeTrack":
	case "playTrack":
		if len(args) > 1{
			setMusicBoxPlaying(args[1],PLAYING) //Set Playing = true
		}
	case "pauseTrack":
		fallthrough //Same as stop
	case "stopTrack":
		if len(args) > 1{
			setMusicBoxPlaying(args[1],PAUSED) //Set Playing = false
		}
	case "nextTrack":
	case "startedTrack":
		log.Print("Track Started")
		//Parse recieved track
		d := data["data"].(map[string]interface{})

		track := d["track"].(map[string]interface{})
		t := TrackItem{
				ProviderID:track["ProviderID"].(string),
				Title:track["Title"].(string),
				ArtistName:track["ArtistName"].(string),
				AlbumName:track["AlbumName"].(string),
				ArtworkURL:track["ArtworkURL"].(string),
				Length:track["Length"].(float64),
		}
		fmt.Println(track)
		deviceID := d["deviceID"].(string)
		fmt.Println(deviceID)


		//
		// Remove from queue
		//

		//popTrackOffQueue(args[1])

		//
		// Save in track history
		//

		//Create aws item
		atts := []dynamodb.Attribute{
			*dynamodb.NewStringAttribute("Title",t.Title),
			*dynamodb.NewStringAttribute("ArtistName",t.ArtistName),
			*dynamodb.NewStringAttribute("ProviderID",t.ProviderID),
			*dynamodb.NewStringAttribute("AlbumName",t.AlbumName), //Moment.us doesn't always provide this
			*dynamodb.NewStringAttribute("ArtworkURL",t.ArtworkURL), //Moment.us doesn't always provide this
			*dynamodb.NewNumericAttribute("Length",strconv.FormatFloat(t.Length,'f',-1,32)), //Moment.us doesn't always provide this
		}

		//Add track to database for this user:musicbox
		if _,err := trackHistoryTable.PutItem(username+":"+deviceID,time.Now().UTC().Format(time.RFC3339),atts); err != nil{
			log.Print(err.Error())
		}else{
			log.Print("Put New track")
		}

		//Set playing to true
		setMusicBoxPlaying(deviceID,PLAYING) //Set Playing = true

	case "updateTheme":
		//Extract ThemeID (All that is needed to find on recommendSongs)
		dataMap := data["data"].(map[string]interface{})
		themeID := dataMap["ThemeID"].(string)

		boxID := args[1]

		//Update box information with new theme
		themeUpdate := []dynamodb.Attribute{*dynamodb.NewStringAttribute("ThemeID",themeID)}

		_, err := musicBoxesTable.UpdateAttributes(&dynamodb.Key{HashKey: boxID},themeUpdate)
		if err != nil{
			log.Print(err.Error())
		}


	default:
			log.Print("Unknown Command:",data["command"])
	}

	return true
}


//
//User Lifecycle
//

func userConnected(authKey string, authExtra map[string]interface{}, permission postmaster.Permissions){
		v,ok := authExtra["client-type"] //Extract client information

		if ok && (v == "musicBox-v1" || v == "testClient-v1"){
			log.Print("Box Connected",authExtra)
			setMusicBoxPlaying(authExtra["client-id"].(string),PAUSED)

			//Send message that device paused
			b,err := lookupMusicBox(authExtra["client-id"].(string))

			if err == nil{
				//Create paused command
				msg := map[string]interface{}{
					"command":"boxConnected",
				}

				server.PublishEvent(baseURL+b.User+"/"+b.ID, msg)
				return
			}
		}else{
		fmt.Println("Connected user: "+authKey)
		}
}

//Called when a websocket is disconnected with the information of the authenticated client
func clientDisconnected(authKey string,authExtra map[string]interface{}){
	v,ok := authExtra["client-type"]

	if ok && (v == "musicBox-v1" || v == "testClient-v1"){
		log.Print("Box Disconnected: ",authExtra)

		setMusicBoxPlaying(authExtra["client-id"].(string),OFFLINE)

		//Send message that device paused
		b,err := lookupMusicBox(authExtra["client-id"].(string))

		if err == nil{
			//Create paused command
			msg := map[string]interface{}{
				"command":"boxDisconnected",
			}

			server.PublishEvent(baseURL+b.User+"/"+b.ID, msg)
		}
	}else{
		log.Print("Client Disconnected: ",authKey)
	}

}
