package main

import(
	"postmaster"
	"fmt"
)


func getUserPremissions(authKey string,authExtra map[string]interface{})(postmaster.Permissions,error){
	
	p := postmaster.Permissions{
		RPC:map[string]postmaster.RPCPermission{
			baseURL+"currentQueueRequest":true,
			baseURL+"players":true,
			baseURL+"recommendSongs":true,
			baseURL+"boxDetails":true,
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

func userConnected(authKey string, authExtra map[string]interface{}, permission postmaster.Permissions){
	fmt.Println("Connected user: "+authKey)
	fmt.Println(authExtra)
}
