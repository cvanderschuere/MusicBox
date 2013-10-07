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
	"errors"
)

//Args [none] (uses conn Username)
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
				boxes = append(boxes,*boxObj) //Only append if exists
			}
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

// Args format {ID:musicBoxID,Count:numToAdd}
func recommendSongs(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){	
	//Get passed options
	opts,ok := args[0].(map[string]interface{})
	if !ok{
		//Incorrect format
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format",Details:""}
	}
	
	//Look up music box with ID
	boxID := opts["ID"].(string)
	requestedCount := int(opts["Count"].(float64))
	
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
	fmt.Println(box.Theme,box.ThemeFull)
	
	//Get theme item for this box
	theme := box.ThemeFull //Already inlcuded in box lookup
	
	fmt.Println(theme)
	
	//Create recommendation request
	v := url.Values{}
	v.Set("access_token", user.SessionID) //Moment.us access token is same as user session id
	
	//
	//Use theme to populate information
	//
	
	//Weather
	if theme.Weather == "AUTO:AUTO"{
		//Find weather information for current location
		
		var condition string
		var temperature string
		
		// TODO: Use actual weather conditions
		//
		// Mock data
		//
		
		condition = "overcast"
		temperature = "20"
		
		//
		// End fillin data
		//
		
		v.Set("by[weather][condition]",condition)	
		v.Set("by[weather][temperature]",temperature)			
	}else{
		//Use provided weather information
		p := strings.Split(theme.Weather,":")
		
		v.Set("by[weather][condition]",p[0])	
		v.Set("by[weather][temperature]",p[1])			
	}
	
	//Time of day []'Early Morning', 'Morning', 'Late Morning', 'Afternoon', 'Late Afternoon', 'Evening', 'Night', 'Late Night'] (required)
	if theme.Time == "AUTO"{
		//Use current time of day to determine value
		var time string
		
		//
		// Mock data
		//
		
		time = "Afternoon"
		//End Mock
		
		v.Set("by[time_of_day]",time)	
	}else{
		v.Set("by[time_of_day]",theme.Time)	
	}
	
	//Mood ['Happy', 'Inspired', 'Tender', 'Nostalgic', 'Relaxed', 'Strong', 'Joyful', 'Tense', 'Sad'] (required)
	v.Set("by[mood]",theme.Mood)
	
	//City (optional)
	if theme.City != "AUTO"{
		v.Set("by[city]",theme.City)	
	}//Else do nothing
	
	//Keywords (optional)
	if len(theme.Keywords)>0{
		v.Set("by[keywords]",strings.Join(theme.Keywords,","))
	}
	
	//Current Context (unchanged by theme)
	v.Set("current_context[date]",time.Now().Format(time.RFC3339))//2006-01-02T15:04:05Z time format layout time.RFC3339
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
			fmt.Println("Json Error: ",err)
			fmt.Println(string(body))
			return []interface{}{}, nil
	}
	
	var tracks []*TrackItem
	c := spotify.New() //Spotify client
	
	Recommend_Loop:
	for _,track := range responseObject.Data{
		//Turn Momemtus track into TrackItem
		t := &TrackItem{}
		t.Title = track.Title
		t.ArtistName = track.Artist.Name
		
		if track.Album.Name != ""{
			t.AlbumName = track.Album.Name
			
		}else{
			t.AlbumName = "UNKNOWN"
		}
		
		
		if len(track.Album.Image) > 0{
			t.ArtworkURL = track.Album.Image[0].Content
		}else{
			t.ArtworkURL = "UNKNOWN"
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
	
	if len(tracks) < requestedCount{
		fmt.Println("Didn't meet request number");
	}
	
	return tracks,nil
}

// Used as a login
//Args [map[username,password]]
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

//Used to get more information about a music box or array of music boxes
//args [ musicBoxId[] ]
func getMusicBoxDetails(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	if len(args) != 1{
		return nil,nil //Not corrent information
	}
	
	idList := args[0].([]interface{})

	boxes := make(map[string]interface{}) //*BoxItem or error
	
	for _,boxID := range idList{
		//Lookup musicbox
		if id,ok := boxID.(string); ok {
			if box,err := lookupMusicBox(id); err == nil{
				//Match with username
				if box.User == conn.Username{
					//Convert box to map
					m := map[string]interface{}{
						"uri":baseURL+conn.Username+"/"+box.ID,
						"box":box,
					}
					
					boxes[id] = m
				
				}else{
					boxes[id] = errors.New("No box exists for this user")
				}
			}else{
				boxes[id] = errors.New("No box exists")
			}
	
		}else{
			boxes[id] = errors.New("Incorrect arg type")
		}
	}
	
	return boxes,nil	
}

//Used to get information about track history
//args [musicboxID returnLimit pivotData(RFC3339)] (pivotDate is such that `returnLimit` items after `pivotDate` are returned)
func getTrackHistory(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	
	var compositeID string
	var limitFloat float64 //Json is float by default
	var limit int //keep into too...more useful
	var date string
	
	if len(args) == 0{
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
	}else{
		//Extract necessary information
		a,ok := args[0].([]interface{})
		if !ok{
			return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (Message)",Details:""}
		}
		
		//Make sure has enough objects
		if len(a)<1{
			return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (Not enough args)",Details:""}
		}
		
		//ID
		compositeID,ok = a[0].(string)
		if !ok{
			return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
		}
		
		//Limit
		if len(a)>1{
			limitFloat,ok = a[1].(float64)
			if !ok{
				return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (limit)",Details:""}
			}
			limit = int(limitFloat)
		}else{
			limit = 0;
		}
		
		//Date
		if len(a)>2{
			//Should do error checking by making sure it converts
			date,ok = a[2].(string)
			if !ok{
				return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (date)",Details:""}
			}
		}else{
			date = time.Now().UTC().Format(time.RFC3339) //Use today as default
		}
	}	
	
	//Query table
	comps := []dynamodb.AttributeComparison{
		*dynamodb.NewEqualStringAttributeComparison("CompositeID",conn.Username+":"+compositeID), //Composite ID
		*dynamodb.NewStringAttributeComparison("Date",dynamodb.COMPARISON_LESS_THAN_OR_EQUAL,date),
	}
	
	res,err := trackHistoryTable.Query(comps)
	
	if err != nil{
		fmt.Println(err)
		return nil,nil
	}else{
		if limit > 0 && limit <= len(res) {
			//limit the return elements to the last couple elements (first in array is oldest)
			res = res[len(res)-limit:]
		}else if limit > len(res){
			res = nil
		}

		tracks := make([]*TrackItem,len(res))
		for i,track := range res{
			tracks[i] = trackItemFromMap(track)
		}
		
		return tracks,nil
	}	
}

//Used to get list of avaliable themes
//Args [none]
func getThemes(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	themes,err := getAllThemes()
	if err != nil{
		return nil, &postmaster.RPCError{URI:uri,Description:err.Error(),Details:""}
	}
	
	return themes,nil
}

