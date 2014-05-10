package main

import(
    "postmaster"
    )

var boxCache map[string]BoxCacheWrapper
var userCache map[string]UserItem

func cachedUserInfoRequest(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){

    if userCache[conn.Username] != nil{
        return userCache[conn.UserName]
    }

    user, err : = userInfoRequest(conn, url, args)

    if err != nil{
        return err
    }

    userCache[conn.Username] = user

    return user, nil
}

func cachedGetQueue(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){
    if len(args) == 0{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (No arguments)",Details:""}
    }

    //Extract necessary information

    //ID
    boxID,ok := args[0].(string)
    if !ok{
        return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
    }

    if boxCache[boxID] != nil{
        return boxCache[boxID].Queue
    }

    tracks,err := getQueue(conn, url, args)

    if err != nil{
        return nil, err
    }

    boxCache[boxID].Queue = tracks

    return tracks, nil
}

func cachedGetTrackHistory(conn *postmaster.Connection,url string, args ...interface{})(interface{},*postmaster.RPCError){
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
        compositeID,ok = a[0].(string)
        if !ok{
            return nil, &postmaster.RPCError{URI:uri,Description:"Invalid format (ID)",Details:""}
        }

        if boxCache[compositeID] != nil{
            return boxCache[compositeID].History
        }

        tracks, err := getTrackHistory(conn, url, args)

        if err != nil{
            return nil, err
        }

        boxCache[compositeID].History = tracks

        return tracks, nil
    }
}


func addTrackToCachedQueue(boxID string, track TrackItem){
    if(boxCache[boxID].Queue == nil){
        return nil
    }

    boxCache[boxID].Queue = append(boxCache[boxID].Queue, TrackItem)
}

func nextTrackInCachedQueue(boxId string, track TrackItem) TrackItem{
    if boxCache[boxID] == nil{
        return nil
    }else{
        if len(boxCache[boxID].Queue) > 0{
                nextTrack := boxCache[boxID].Queue[0]

                if nextTrack == track{
                    boxCache[boxID].Queue = boxCache[boxID].Queue[1:]

                    boxCache[boxID].History = append(boxCache[boxId].History, nextTrack)

                    return nextTrack
                }else{
                    return nil
                }
        }else{
            return nil
        }
    }

}
