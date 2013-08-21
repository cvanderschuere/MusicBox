package main

import(
	"net/http"
	"net/url"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"strings"
)


//Find `numToAdd` songs similar to `baseTrack` and send addTrack message with them
func addSimilarSongs(baseTrack MusicBoxTrack, numToAdd uint){
	
	//Launch webrequest for similar songs
	songsToAdd := findSimilarSongsLastFM(baseTrack,numToAdd)
	
	//Send addTrack to all devices (including self) with new songs
	data := make([](map[string]string),len(songsToAdd))
	for i,song := range songsToAdd{
		songDict := map[string]string{"trackName":song.TrackName, "albumName":song.AlbumName, "artistName":song.ArtistName, "service":song.Service, "url":song.URL}
		data[i] = songDict
	}
	
	addMsg := map[string]interface{}{"command":"addTrack", "data":data}
	client.Publish(baseURL+username+"/"+deviceName,addMsg) 
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

//Use Last.fm's track.getSimilar API to find songs
func findSimilarSongsLastFM(baseTrack MusicBoxTrack, numToAdd uint)([]MusicBoxTrack){
	
	//Make track.getsimilar request (need at least 2 to force array return type & should give padding in case of unfound songs)
	lastFMURL := baseURLLastFM + 
			fmt.Sprintf("&artist=%s&track=%s&limit=%d",url.QueryEscape(baseTrack.ArtistName),url.QueryEscape(baseTrack.TrackName),(numToAdd+1)*2)
	
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
	}
				
	returnChan := make(chan MusicBoxTrack,numToAdd)	
		
	//Convert to MusicBoxTracks (match to spotify)
	for _,similarTrack := range responseObject.Similartracks.Track{
		//Start concurrent update loop
		go matchToSpotify(similarTrack,returnChan)
	}
	
	var addedTracks []MusicBoxTrack
	
	//Read in created tracks
	for uint(len(addedTracks))<numToAdd{
		newTrack := <-returnChan
		addedTracks = append(addedTracks,newTrack)
	}
	
	return addedTracks
}

//Make search call to spotify with artist name & track name 
//Return musicBoxTrack on chan if found
func matchToSpotify(track lastFMTrack,resultChan chan MusicBoxTrack){
	response,error := http.Get(baseURLSpotifySearch + url.QueryEscape(track.Artist.Name+" "+track.Name))
	if error != nil {
		// handle error
		log.Error("Spotify Search Error:%s",error)
		return
	}

	bodySpotify, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	//Unmarshal JSON
	var responseObject spotifyTrackSearchResponse
	err := json.Unmarshal(bodySpotify,&responseObject)
	if err != nil {
		log.Error("Json Error(Spotify): %s",err)
		return
	}
	
	if len(responseObject.Tracks) == 0{
		log.Error("Spotify found no matching tracks:",track.Artist.Name,track.Name)
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
	newTrack := MusicBoxTrack{AlbumName:foundSpotifyTrack.Album.Name,ArtistName:foundSpotifyTrack.Artists[0].Name,TrackName:foundSpotifyTrack.Name,Service:"Spotify",URL:foundSpotifyTrack.Href}
	resultChan<-newTrack
}