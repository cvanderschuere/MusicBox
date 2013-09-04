package main

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
	CompositeID	string //username:BoxID
	Date	string  //Date played for accounting purposes
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