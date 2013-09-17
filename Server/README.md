API Documentation
=================

Base URL: http://www.musicbox.com/

RPC
---

### Authorized
*	recommendSongs
*	players
	-	args [none] (uses conn Username)
	-	return: [musicBoxID]
*	boxDetails
	-	args: [ musicBoxId[] ]
	-	map[ musicboxID:map[DeviceName,ID,Theme,Location[],Theme,User] ]
*	trackHistory
	-	args: [musicboxID returnLimit pivotData(RFC3339)] (pivotDate is such that `returnLimit` items after `pivotDate` are returned)
	-	return: [ map[AlbumName,ArtistName,ArtworkURL,Date,ProviderID,Title] ]

### Unauthorized
*	user/startSession
	-	args:[map{username,password}]
	-	return:map{username,sessionID}
*	musicbox/startSession
