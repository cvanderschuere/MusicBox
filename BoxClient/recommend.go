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

//Echonest
type echoSongSearchResponse struct{
	Response songSearchBody
}
type songSearchBody struct{
	Songs []matchedSong
}
type matchedSong struct{
	Tracks []foreignSongID
}
type foreignSongID struct{
	Foreign_id string
}

//Spotify
type spotifyTrackLookupResponse struct{
	Track spotifyLookupTrack
}
type spotifyLookupTrack struct{
	Name string
	Artists []spotifyLookupArtist
	Album spotifyLookupAlbum
}
type spotifyLookupArtist struct{
	Name string
}
type spotifyLookupAlbum struct{
	Name string
}

const baseURLLastFM = "http://ws.audioscrobbler.com/2.0/?method=track.getsimilar&api_key=600be92e4856b530ec9ffaef2906e5a6&format=json" 
const baseURLEchoNest = "http://developer.echonest.com/api/v4/song/search?api_key=MRVCCYJZYJ32THKA8&format=json&results=1&bucket=id:spotify-WW&bucket=tracks"
const baseURLSpotifyLookup = "http://ws.spotify.com/lookup/1/.json?uri="

//Use Last.fm's track.getSimilar API to find songs
func findSimilarSongsLastFM(baseTrack MusicBoxTrack, numToAdd uint)([]MusicBoxTrack){
	
	//Make track.getsimilar request (need at least 2 to force array return type & should give padding in case of unfound songs)
	lastFMURL := baseURLLastFM + 
			fmt.Sprintf("&artist=%s&track=%s&limit=%d",url.QueryEscape(baseTrack.ArtistName),url.QueryEscape(baseTrack.TrackName),(numToAdd+1)*5)
	
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
		go func(){
			//Make EchoNest call to match artistName & trackName to spotifyID
			echoURL := baseURLEchoNest + fmt.Sprintf("&artist=%s&title=%s",url.QueryEscape(similarTrack.Artist.Name),url.QueryEscape(similarTrack.Name))
			response,error := http.Get(echoURL)
			if error != nil {
				// handle error
				log.Error("EchoNest Error:%s",error)
				return
			}
		
			bodyEcho, _ := ioutil.ReadAll(response.Body)
			response.Body.Close()
	
			//Unmarshal JSON
			var responseObjectEcho echoSongSearchResponse
			err = json.Unmarshal(bodyEcho,&responseObjectEcho)
			if err != nil {
				log.Error("Json Error(Echo): %s",err)
				return
			}
		
			if len(responseObjectEcho.Response.Songs) == 0  || len(responseObjectEcho.Response.Songs[0].Tracks) == 0{
				//Failed to find match
				log.Error("Echo Nest Failed to find spotify URI")
				return
			}
		
			//Extract spotify uri
			spotifyURLEcho := responseObjectEcho.Response.Songs[0].Tracks[0].Foreign_id
			spotifyURL := strings.Replace(spotifyURLEcho,"spotify-WW","spotify",1)
		
			//Lookup spotify track for detailed info (track,artist,album)
			response,error = http.Get(baseURLSpotifyLookup+spotifyURL)
			bodySpotfy, _ := ioutil.ReadAll(response.Body)
			response.Body.Close()		
		
			//Unmarshal JSON
			var responseSpotify spotifyTrackLookupResponse
			err = json.Unmarshal(bodySpotfy,&responseSpotify)
			if err != nil {
				log.Error("Json Error(Spotify): %s",err)
				return
			}
		
			foundSpotifyTrack := responseSpotify.Track
		
			//Form musicBoxTrack and pass on channel
			newTrack := MusicBoxTrack{AlbumName:foundSpotifyTrack.Album.Name,ArtistName:foundSpotifyTrack.Artists[0].Name,TrackName:foundSpotifyTrack.Name,Service:"Spotify",URL:spotifyURL}
			returnChan<-newTrack
		}()
	}
	
	var addedTracks []MusicBoxTrack
	
	//Read in created tracks
	for uint(len(addedTracks))<numToAdd{
		newTrack := <-returnChan
		
		addedTracks = append(addedTracks,newTrack)
		
	}
	
	
	return addedTracks
}