package main

import(
    "postmaster"
    )

var boxCache map[string]BoxCacheWrapper
var userCache map[string]UserItem

func cachedUserInfoRequest(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){

    if user, ok := userCache[conn.Username]; ok{
        return user, nil
    }

    user, err := userInfoRequest(conn, url, args)

    if err != nil{
        return nil, err
    }

    userCache[conn.Username] = user.(UserItem)

    return user, nil
}

func cachedGetQueue(conn *postmaster.Connection,uri string, args ...interface{})(interface{},*postmaster.RPCError){
    if len(args) == 0{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
    }

    //Extract necessary information

    //ID
    boxID,ok := args[0].(string)
    if !ok{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
    }

    if box, ok := boxCache[boxID]; ok && len(box.Queue) > 0{
        return box.Queue, nil
    }else if !ok {
        boxCache[boxID] = *new(BoxCacheWrapper)
    }

    tracks,err := getQueue(conn, uri, args)

    if err != nil{
        return nil, err
    }

    box := boxCache[boxID]
    box.Queue = tracks.([]TrackItem)
    boxCache[boxID] = box

    return tracks, nil
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

        tracks, err := getTrackHistory(conn, uri, args)

        if err != nil{
            return nil, err
        }

        box := boxCache[compositeID]
        box.History = tracks.([]TrackItem)
        boxCache[compositeID] = box

        return tracks, nil
    }
}


func addTrackToCachedQueue(boxID string, track TrackItem){
    if _, ok := boxCache[boxID]; !ok{
	boxCache[boxID] = *new(BoxCacheWrapper)
    }

    box := boxCache[boxID]
    box.Queue = append(boxCache[boxID].Queue, track)
    boxCache[boxID] = box
}

func nextTrackInCachedQueue(boxID string, track TrackItem) TrackItem{
    if box, ok := boxCache[boxID]; !ok{
        boxCache[boxID] = *new(BoxCacheWrapper)
    }else{
        if len(box.Queue) > 0{
	    nextTrack := box.Queue[0]

	    if nextTrack == track{
	        box.Queue = boxCache[boxID].Queue[1:]

	        box.History = append(boxCache[boxID].History, nextTrack)

		boxCache[boxID] = box

	    	return nextTrack
	    }
        }else{
            box.History = append(boxCache[boxID].History, track)

	    boxCache[boxID] = box
        }
    }
    
    return track
}
