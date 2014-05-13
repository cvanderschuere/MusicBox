package main

import(
    "postmaster"
    "fmt"
    )

var boxCache map[string]BoxCacheWrapper
var userCache map[string]*UserItem

func setupCaches(){
    boxCache = make(map[string]BoxCacheWrapper, 10)
    userCache = make(map[string]*UserItem, 5)
}

func cachedUserInfoRequest(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){

    if user, ok := userCache[conn.Username]; ok{
        return user, nil
    }

    user, err := userInfoRequest(conn, url, args)

    if err != nil{
        return nil, err
    }

    userCache[conn.Username] = user.(*UserItem)

    return user, nil
}

func cachedGetQueue(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
    if len(args) == 0{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
    }

	//Extract necessary information
	boxID,ok := args[0].(string)
    if !ok{
	fmt.Println(ok)
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
    }

    if box, ok := boxCache[boxID]; ok && len(box.Queue) > 0{
fmt.Println("return queue: ", box.Queue)
        return box.Queue, nil
    }
fmt.Println("create Queue")
	box := *new(BoxCacheWrapper)

	box.Queue = make([]TrackItem, 0)
	boxCache[boxID] = box

	return box.Queue, nil
}

func cachedGetTrackHistory(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
    if len(args) == 0{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
    }else{
        //Extract necessary information
        a,ok := args[0].([]interface{})
        if !ok{
            return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (Message)",Details:""}
        }

        //Make sure has enough objects
        if len(a)<1{
            return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (Not enough args)",Details:""}
        }

        //ID
        compositeID,ok := a[0].(string)
        if !ok{
            return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
        }

        if box,ok := boxCache[compositeID]; ok && len(box.History) > 0{
            return box.History, nil
        }else if !ok {
            boxCache[compositeID] = *new(BoxCacheWrapper)
        }

        tracks, err := getTrackHistory(conn, uri, args[0])

        if err != nil{
            return nil, err
        }

        box := boxCache[compositeID]
        box.History = tracks.([]*TrackItem)
        boxCache[compositeID] = box

        return tracks, nil
    }
}


func addTrackToCachedQueue(boxID string, track TrackItem){
fmt.Println("add track to cached queue:", track)
    if _, ok := boxCache[boxID]; !ok{
	box := *new(BoxCacheWrapper)

	box.Queue = make([]TrackItem, 0)
	boxCache[boxID] = box
    }

    box := boxCache[boxID]
    box.Queue = append(boxCache[boxID].Queue, track)
    boxCache[boxID] = box
}

func nextTrackInCachedQueue(boxID string, track TrackItem) TrackItem{
fmt.Println("next track in cached queue:", track)
    if box, ok := boxCache[boxID]; !ok{
	box := *new(BoxCacheWrapper)

	box.Queue = make([]TrackItem, 0)
	boxCache[boxID] = box
    }else{
        if len(box.Queue) > 0{
	    nextTrack := box.Queue[0]

	    if nextTrack.Title == track.Title{
	        box.Queue = boxCache[boxID].Queue[1:]

	        box.History = append(boxCache[boxID].History, &nextTrack)

		boxCache[boxID] = box

	    	return nextTrack
	    }
        }else{
            box.History = append(boxCache[boxID].History, &track)

	    boxCache[boxID] = box
        }
    }
    
    return track
}
