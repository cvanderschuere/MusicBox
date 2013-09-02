package main

type UserItem struct{
	Username string
	Password string `json:",omitempty"` //Allows use of same struct for send/recieve
	SessionID string
	
	MusicBoxes []string
}