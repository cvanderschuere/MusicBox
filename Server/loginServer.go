//Simple package that provides interface to users database
package main

import(
	"log"
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"	
)

//Globals
var dynamoDBServer *dynamodb.Server
var usersTable *dynamodb.Table

func main() {
	
	if err := setupAWS();err != nil{
		log.Fatal("AWS Login Error: err")
		return
	}
	
	http.HandleFunc("/login", loginHandler)
	
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

//
// Handler
//

func loginHandler(w http.ResponseWriter, req *http.Request) {
	
	//Get header information to auth request
	username := req.Header.Get("musicbox-username")
	password := req.Header.Get("musicbox-password")
	
	userObj := &UserItem{}
	
	//Verify if accurate information
	if username != "" && password != ""{
		
		//Get user information from database
		if item, err := usersTable.GetItem(&dynamodb.Key{HashKey: username}); err == nil{

			err := dynamodb.UnmarshalAttributes(&item, userObj)
			if err != nil {
				http.Error(w,"Error Processing Request",500)
				return
			}
			
			//Compare session id to verify accuracy
			if userObj.Password != password{
				http.Error(w,"Incorrect Password",401)
				return
			}			
		}
	}
	
	//Write out user information
	w.Header().Set("Content-Type", "application/json")
	
	
	//Clear sensitive information
	userObj.Password = ""
	
	json,err := json.Marshal(userObj)
	if err != nil {
		http.Error(w,"Error Processing Request",500)
		return
	}
	
	w.Write(json)
}

//
//	Amazon Web Services
//

func setupAWS()(error){
	//Sign in to AWS
	auth, err := aws.EnvAuth()

	if err != nil {
		return fmt.Errorf("Signin to AWS failed: %s",err.Error())
	}

	dynamoDBServer = &dynamodb.Server{auth, aws.USWest2}
	
	//Create tables
	
	//Users
	primary := dynamodb.NewStringAttribute("username", "")
	key := dynamodb.PrimaryKey{primary, nil}
	usersTable = dynamoDBServer.NewTable("Users",key)	

	return nil
}

