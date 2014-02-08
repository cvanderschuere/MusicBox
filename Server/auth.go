package main

import(
	"postmaster"
)


func getUserPremissions(authKey string,authExtra map[string]interface{})(postmaster.Permissions,error){
	
	p := postmaster.Permissions{
		RPC:map[string]postmaster.RPCPermission{
			baseURL+"players":true,
			baseURL+"recommendSongs":true,
			baseURL+"userInfo":true,
			baseURL+"boxDetails":true,
			baseURL+"trackHistory":true,
			baseURL+"queue":true,
		},
		PubSub:map[string]postmaster.PubSubPermission{
		},
	}
	user,err := lookupUser(authKey)
	if err == nil{
		//Add pubSub for all music boxes [base+username+boxid]
		for _,boxID := range user.MusicBoxes{
			p.PubSub[baseURL+authKey+"/"+boxID] = postmaster.PubSubPermission{true,true}
		}
	}
	
	return p,nil
}
