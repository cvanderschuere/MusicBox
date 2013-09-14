package main

import(
	"github.com/crowdmob/goamz/dynamodb"
	"fmt"
)

type UserItem struct{
	Username string
	Password string `json:",omitempty"` //Allows use of same struct for send/recieve
	SessionID string
	
	MusicBoxes []string
}

type BoxItem struct{
	ID	string
	User string
	DeviceName string
	Location	[]string
	Theme	string
	
	//Dynamic stats
	Playing 	int64
}

type TrackItem struct{
	Title string
	ArtistName string
	AlbumName	string
	ArtworkURL	string
	
	//Track info
	ProviderID	string
	
	//Storage info
	CompositeID	string `json:",omitempty"` //username:BoxID
	Date	string  //Date played for accounting purposes
}

func trackItemFromMap(data map[string]*dynamodb.Attribute)(*TrackItem){
	fmt.Println(data)

	t := &TrackItem{
		Title: data["Title"].Value,
		ArtistName: data["ArtistName"].Value,
		//AlbumName: data["AlbumName"].Value,
		//ArtworkURL: data["ArtworkURL"].Value,
		ProviderID: data["ProviderID"].Value,
		Date: data["Date"].Value,
	}
	
	return t
}

// Moment.us
type DiscoverResult struct{
	Data []MomentusTrack
}

type MomentusTrack struct{
	Title string
	
	//Links
	Artist MomentusArtist
	Album MomentusAlbum
}
type MomentusArtist struct{
	Name string
	Image []MomentusImage
}
type MomentusAlbum struct{
	Name string
	Image []MomentusImage
}
type MomentusImage struct{
	Size string
	Content string
}
