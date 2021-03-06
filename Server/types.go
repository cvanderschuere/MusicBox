package main

import(
	"github.com/crowdmob/goamz/dynamodb"
	"strconv"
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
	ThemeFull *ThemeItem
	
	//Dynamic stats
	Playing 	int64
}

const ( 
		OFFLINE int64 = iota
    	PAUSED int64 = iota 
        PLAYING int64 = iota 
)

type TrackItem struct{
	//Generic Information
	Title string
	ArtistName string
	AlbumName	string
	ArtworkURL	string
	Length	float64
	
	//Track info
	ProviderID	string
	
	//Storage info
	CompositeID	string `json:",omitempty"` //username:BoxID
	Date	string  //Date played for accounting purposes
}

func trackItemFromMap(data map[string]*dynamodb.Attribute)(*TrackItem){
	
	//Extract length
	l,_ := strconv.ParseFloat(data["Length"].Value,32)

	t := &TrackItem{
		Title: data["Title"].Value,
		ArtistName: data["ArtistName"].Value,
		AlbumName: data["AlbumName"].Value,
		ArtworkURL: data["ArtworkURL"].Value,
		ProviderID: data["ProviderID"].Value,
		Length: l,
		Date: data["Date"].Value,
	}
	
	return t
}

type ThemeItem struct{
	ThemeID string
	Name	string
	
	//Moment.us specific
	City string
	DayOfWeek string
	Keywords []string
	Mood string
	Time string
	Weather string	
}

func themeItemFromMap(data map[string]interface{})(*ThemeItem){
	t := &ThemeItem{
		ThemeID: data["ThemeID"].(string),
		Name: data["Name"].(string),
		
		City: data["City"].(string),
		DayOfWeek: data["DayOfWeek"].(string),
		Mood: data["Mood"].(string),
		Time: data["Time"].(string),
		Weather: data["Weather"].(string),
	}
	
	//Cast keywords
	keys := data["Keywords"].([]interface{})
	keysNew := make([]string,len(keys))
	for i,key := range keys{
		keysNew[i] = key.(string)
	}
	t.Keywords = keysNew
	
	
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
