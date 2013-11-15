package Track

import()

//Fields must be exported for JSON marshal
type Track struct{
	Title string
	ArtistName string
	AlbumName	string
	ArtworkURL	string
	Length	float64
	
	//Track info
	ProviderID	string
	
	//Storage info
	CompositeID	string //username:BoxID
	Date	string  //Date played for accounting purposes
}


func FromMap(data map[string]interface{})(Track){
	
	//Extract length
	//l,_ := strconv.ParseFloat(data["Length"].(string),32)

	t := Track{
		Title: data["Title"].(string),
		ArtistName: data["ArtistName"].(string),
		AlbumName: data["AlbumName"].(string),
		ArtworkURL: data["ArtworkURL"].(string),
		ProviderID: data["ProviderID"].(string),
		Length: data["Length"].(float64),
		Date: data["Date"].(string),
	}
	
	return t
}
