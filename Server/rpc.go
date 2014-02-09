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
	"github.com/cvanderschuere/go-pandora"
)

//Args [none] (uses conn Username)
func userInfoRequest(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){
	user,err2 := lookupUser(conn.Username)
	
	//Clear sensitive info
	user.Password = ""
	
	if err2 != nil{
		return nil,err2
	}
	
	return user,nil
}

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

	//See if we can get them from the queue
	queue,errQ := getQueueTracks(boxID,requestedCount)
	if len(queue)==requestedCount && errQ == nil{
		return queue,nil
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
	
	//Determine what type of theme this is
	if box.ThemeFull.Type == PandoraTheme{
		returnMap := map[string]interface{}{
			"type":PandoraTheme,
			"id":box.ThemeFull.ThemeID,
		}
		
		return returnMap,nil
		
	}else if box.ThemeFull.Type == MomentusTheme{
		//Make Moment.us request based on box information
		fmt.Println(box.ThemeID,box.ThemeFull)

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
	
		var tracks []TrackItem
		c := spotify.New() //Spotify client

		Recommend_Loop:
		for _,track := range responseObject.Data{
			//Turn Momemtus track into TrackItem
			t := TrackItem{}
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
							//Fill in Spotify Specific Information

							t.ProviderID = track.URI
							t.Length = track.Length

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
	
		//Protect against nil queue
		if queue != nil && len(queue)>0{
			tracks = append(queue,tracks...)
		}
	
		if len(tracks) < requestedCount{
			fmt.Println("Recommendation Error")
			return nil, &postmaster.RPCError{URI:uri,Description:"Recommendation Error",Details:""}
		}
		
		returnMap := map[string]interface{}{
			"type":MomentusTheme,
			"id":box.ThemeFull.ThemeID,
			"track":tracks[:requestedCount],
		}

		return returnMap,nil
	}else{
		//Unknown theme type
		return nil,&postmaster.RPCError{URI:uri,Description:"Unknown Theme",Details:""}
	}
}

// Used as a login
//Args [map[username,password]]
func startSession(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	searchUser := args[0].(map[string]interface{})

	user,err := lookupUser(strings.ToLower(searchUser["username"].(string)))
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
		err = &postmaster.RPCError{URI:"",Description:"Invalid Password",Details:""}
		return nil,err
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
		}

		tracks := make([]*TrackItem,len(res))
		for i,track := range res{
			tracks[i] = trackItemFromMap(track)
		}

		return tracks,nil
	}
}

//Used to get queue information for given device
//Args musicBoxID
//Returns: [queueItem1 queueItem2]
func getQueue(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	//Check for valid input
	if len(args) == 0{
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
	}
	
	//Extract necessary information
	
	//ID
	boxID,ok := args[0].(string)
	if !ok{
		return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
	}
	
	//Count must be 1-10
	if tracks,err := getQueueTracks(boxID,10); err != nil{
		fmt.Println(err)
		return nil,&postmaster.RPCError{URI:uri,Description:err.Error(),Details:""}
	}else{
		fmt.Println("Returned Queue")
		return tracks,nil		
	}	
}

//Used to get list of avaliable themes
//Args [none]
func getThemes(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
	
	//Get Pandora themes in parallel
	returnChan := make(chan []*ThemeItem,1)
	go func(c chan []*ThemeItem, username string){
		//Lookup the user
		user,err := lookupUser(username)
		if err != nil{
			c<-nil
			return
		}
		
		//Login to pandora
		client,login := pandora.Login(username,user.PandoraPassword)
		loginError := <-login

		//Probably wrong username/password
		if loginError != nil{
			c<-nil
			return
		}

		stations,stationError := client.GetStationList()

		if stationError != nil || len(stations) == 0{
			c<-nil
			return
		}
		
		//Convert pandoraStation into ThemeItem
		themeList := make([]*ThemeItem,len(stations))
		for i,station := range stations{
			themeList[i] = &ThemeItem{ThemeID:station.ID,Name:station.Name,ArtworkURL:station.ArtworkURL,Type:PandoraTheme}
		}

		c<-themeList
		
		//Try Logout
		logout := client.Logout()
		<-logout
		
	}(returnChan,conn.Username)
	
	//Get moment.us themes
	/*
	//Get AWS Themes
	themes,err := getAWSThemes()
	if err != nil{
		return nil, &postmaster.RPCError{URI:uri,Description:err.Error(),Details:""}
	}
	
	
	//Merge the two
	themes = append(themes,pandoraThemes...)
	*/
	pandoraThemes := <-returnChan
	

	return pandoraThemes,nil
}

