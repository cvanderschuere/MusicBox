package MusicPlayer

import(
	// "net/http"
	// "net/url"
	// "fmt"
	// "encoding/json"
	// "io/ioutil"
	// "strings"
	// "MusicBox/BoxClient/Track"
)

//Find `numToAdd` songs similar to `baseTrack` and send addTrack message with them
func recommendSongs(numToAdd uint){
	/*
	opts := map[string]interface{}{
		"ID":musicBoxID,
		"Count":5,
	}
	
	
	if err := client.Call("recommendSongs",baseURL+"recommendSongs",opts); err != nil{
		fmt.Println(err.Error())
	}
	*/
	//Return call handled with other events
}

//JSON structs 

//Last.fm
type getSimilarResponse struct{
	Similartracks similarTrackResponse
}
type similarTrackResponse struct{
	Track []lastFMTrack
}
type lastFMTrack struct{
	Artist lastFMArtist
	Name string
	Match string
	Mbid string
}
type lastFMArtist struct{
	Mbid string
	Name string
}

//Spotify
type spotifyTrackSearchResponse struct{
	Tracks []spotifySearchTrack
}
type spotifySearchTrack struct{
	Name string
	Href string
	Artists []spotifySearchArtist
	Album spotifySearchAlbum
}
type spotifySearchArtist struct{
	Name string
	Href string
}
type spotifySearchAlbum struct{
	Name string
	Href string
	Availability spotifySearchAvailability
}
type spotifySearchAvailability struct{
	Territories string
}

const baseURLLastFM = "http://ws.audioscrobbler.com/2.0/?method=track.getsimilar&api_key=600be92e4856b530ec9ffaef2906e5a6&format=json" 
const baseURLSpotifySearch = "http://ws.spotify.com/search/1/track.json?q="

/*
//Use Last.fm's track.getSimilar API to find songs
func findSimilarSongsLastFM(baseTrack Track.Track, numToAdd uint)([]Track.Track){
	
	//Make track.getsimilar request (need at least 2 to force array return type & should give padding in case of unfound songs)
	lastFMURL := baseURLLastFM + 
			fmt.Sprintf("&artist=%s&track=%s&limit=%d",url.QueryEscape(baseTrack.ArtistName),url.QueryEscape(baseTrack.Title),(numToAdd)*5)
	
	log.Debug(lastFMURL)
	resp,err := http.Get(lastFMURL)
	if err != nil {
		// handle error
		log.Error("Last.fm Error:%s",err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
		
	//Unmarshal JSON
	var responseObject getSimilarResponse
	err = json.Unmarshal(body,&responseObject)
	if err != nil {
			log.Error("Json Error: %s",err)
			log.Trace(string(body))
	}
				
	returnChan := make(chan *Track.Track,numToAdd)	
		
	//Convert to MusicBoxTracks (match to spotify)
	for _,similarTrack := range responseObject.Similartracks.Track{
		
		//Start concurrent update loop
		go matchToSpotify(similarTrack,returnChan)
	}
	
	var addedTracks []Track.Track
	
	//Read in created tracks
	for i:=0;i<len(responseObject.Similartracks.Track);i++{
		newTrack := <-returnChan
		if newTrack != nil{
			addedTracks = append(addedTracks,*newTrack)
		}
	}
		
	if len(addedTracks)>int(numToAdd){
		addedTracks = addedTracks[:numToAdd] //Pass only last numToAdd songs
	}
	
	return addedTracks
}

//Make search call to spotify with artist name & track name 
//Return musicBoxTrack on chan if found
func matchToSpotify(track lastFMTrack,resultChan chan *Track.Track){
	response,error := http.Get(baseURLSpotifySearch + url.QueryEscape(track.Artist.Name+" "+track.Name))
	if error != nil {
		// handle error
		log.Error("Spotify Search Error:%s",error)
		resultChan<-nil
		return
	}

	bodySpotify, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	//Unmarshal JSON
	var responseObject spotifyTrackSearchResponse
	err := json.Unmarshal(bodySpotify,&responseObject)
	if err != nil {
		log.Error("Json Error(Spotify): %s",err)
		resultChan<-nil
		return
	}
	
	if len(responseObject.Tracks) == 0{
		log.Error("Spotify found no matching tracks:",track.Artist.Name,track.Name)
		resultChan<-nil
		return
	}
	
	//Make assumption that top track returned by spotify is correct one

	//Find track that is avaliable in us (making assumption that spotify returned a proper track)
	var foundSpotifyTrack spotifySearchTrack
	for _,spotifyTrack := range responseObject.Tracks{
		if strings.Contains(spotifyTrack.Album.Availability.Territories,"US"){
			foundSpotifyTrack = spotifyTrack
			break
		}
	}

	//Form musicBoxTrack and pass on channel(sync)
	newTrack := &Track.Track{AlbumName:foundSpotifyTrack.Album.Name,ArtistName:foundSpotifyTrack.Artists[0].Name,Title:foundSpotifyTrack.Name,ProviderID:foundSpotifyTrack.Href}
	resultChan<-newTrack
}
*/
