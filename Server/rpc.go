package main

import(
	"postmaster"
	"github.com/iand/spotify"
	"encoding/json"
	"io/ioutil"
	"sort"
	"fmt"
	"github.com/crowdmob/goamz/dynamodb"
	"net/url"
	"time"
	"net/http"
	"strings"
)

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
