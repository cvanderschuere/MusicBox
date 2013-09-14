API Documentation
=================

Base URL: http://www.musicbox.com/

RPC
---

### Authorized
*	recommendSongs
*	players
	-	args [none] (uses conn Username)
*	boxDetails
	-	args: [ musicBoxId[] ]
*	trackHistory
	-	args: [musicboxID returnLimit pivotData(RFC3339)] (pivotDate is such that `returnLimit` items after `pivotDate` are returned)
	
### Unauthorized
*	user/startSession
*	musicbox/startSession
