package main

import(
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"errors"
	"postmaster"
	"fmt"
	"strconv"
)

//Global database tables (setup in setupAWS())
var dynamoDBServer *dynamodb.Server
var usersTable *dynamodb.Table
var musicBoxesTable *dynamodb.Table
var trackHistoryTable *dynamodb.Table
var themesTable *dynamodb.Table

func setupAWS()(error){
	//Sign in to AWS
	auth, err := aws.EnvAuth()

	if err != nil {
		return errors.New("Signin to AWS failed: "+err.Error())
	}

	dynamoDBServer = &dynamodb.Server{auth, aws.USWest2}

	//Users
	primary := dynamodb.NewStringAttribute("Username", "")
	key := dynamodb.PrimaryKey{primary, nil}
	usersTable = dynamoDBServer.NewTable("Users",key)

	//MusicBoxes
	primary = dynamodb.NewStringAttribute("ID", "")
	key = dynamodb.PrimaryKey{primary, nil}
	musicBoxesTable = dynamoDBServer.NewTable("MusicBoxes",key)

	//Track History
	primary = dynamodb.NewStringAttribute("CompositeID", "")
	rangeKey := dynamodb.NewStringAttribute("Date", "")
	key = dynamodb.PrimaryKey{primary, rangeKey}
	trackHistoryTable = dynamoDBServer.NewTable("TrackHistory",key)

	//Themes
	primary = dynamodb.NewStringAttribute("ThemeID", "")
	key = dynamodb.PrimaryKey{primary, nil}
	themesTable = dynamoDBServer.NewTable("Themes",key)

	return nil
}

//
// Item lookup
//

func lookupMusicBox(id string)(*BoxItem,*postmaster.RPCError){
	boxObj := &BoxItem{}

	//Get music box
	if box, err3 := musicBoxesTable.GetItem(&dynamodb.Key{HashKey: id}); err3 == nil{
		boxErr := dynamodb.UnmarshalAttributes(&box, boxObj)
		if boxErr != nil {
			boxErr2 := &postmaster.RPCError{URI:"",Description:"Unmarshal Error",Details:""}
			return nil,boxErr2
		}else{
			//Fill in theme information
			themeObj := &ThemeItem{}

			theme, _ := themesTable.GetItem(&dynamodb.Key{HashKey:box["ThemeID"].Value})
			dynamodb.UnmarshalAttributes(&theme, themeObj)

			boxObj.ThemeFull = themeObj;

			return boxObj,nil
		}
	}else{
		err2 := &postmaster.RPCError{URI:"",Description:"Invalid BoxID",Details:""}
		return nil, err2
	}
}

func lookupUser(username string)(*UserItem,*postmaster.RPCError){
	if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: username}); err == nil{
		userObj := &UserItem{}

		err := dynamodb.UnmarshalAttributes(&item, userObj)
		if err != nil {
			err2 := &postmaster.RPCError{URI:"",Description:"Unmarshal Error",Details:""}
			return nil,err2
		}else{
			return userObj,nil
		}
	}else{
		err2 := &postmaster.RPCError{URI:"",Description:"Invalid user",Details:""}
		return nil, err2
	}
}

func lookupUserSessionID(authKey string)(string,error){
	fmt.Println("Lookup user: "+authKey);

	user,err := lookupUser(authKey)
	if err == nil{
		return user.SessionID,nil
	}else{
		return "",errors.New("Can't find user")
	}
}
func lookupTheme(themeID string)(*ThemeItem,error){
	if item, err := themesTable.GetItem(&dynamodb.Key{HashKey: themeID}); err == nil{
		themeObj := &ThemeItem{}

		err := dynamodb.UnmarshalAttributes(&item, themeObj)
		if err != nil {
			return nil,err
		}else{
			return themeObj,nil
		}
	}else{
		return nil, err
	}
}

func getAllThemes()([]*ThemeItem,error){
	//Scan table for all items
	res,_ := themesTable.Scan(nil)

	themes := make([]*ThemeItem,len(res))
	for i,val := range res{
		theme := &ThemeItem{}
		dynamodb.UnmarshalAttributes(&val,theme)

		themes[i] = theme
	}


	return themes,nil
}

//
// Update item
//

func setMusicBoxPlaying(musicBoxID string, playing int64)(error){

	update := []dynamodb.Attribute{*dynamodb.NewNumericAttribute("Playing",strconv.FormatInt(playing,10))}
	musicBoxesTable.UpdateAttributes(&dynamodb.Key{HashKey: musicBoxID},update)

	return nil

}