package main

import(
	"code.google.com/p/go.crypto/pbkdf2"
	"encoding/base64"
	"crypto/sha256"
	"crypto/hmac"
)

//
// Crypto
//

/*
	Computes a derived cryptographic key from a password according to PBKDF2 http://en.wikipedia.org/wiki/PBKDF2.

	The function will only return a derived key if at least 'salt' is present in the 'extra' dictionary. The complete set of attributesthat can be set in 'extra':

         salt: The salt value to be used.
         iterations: Number of iterations of derivation algorithm to run.
         keylen: Key length to derive.

	returns the derived key or the original secret.
*/
func deriveKey(secret string, extra map[string]interface{})(string){
	//Salt needed to derive key
	if salt,ok := extra["salt"]; ok{
		iter := 10000
		keyLen := 32
		
		//Check for custom values
		if cIter,ok := extra["iterations"]; ok{
			iter = cIter.(int)
		}
		if cKeylen,ok := extra["keylen"];ok{
			keyLen = cKeylen.(int)
		}
		
		dk := pbkdf2.Key([]byte(secret), []byte(salt.(string)), iter, keyLen, sha256.New)
		key := base64.StdEncoding.EncodeToString(dk)
		
		return key
		
	}else{
		//just return secret
		return secret
	}	
}

func authSignature(authChallenge []byte,authSecret string, authExtra map[string]interface{})(string){
	//Derive authsecret
	authSecret = deriveKey(authSecret,authExtra)
	
	//Generate signature
	sHash :=  hmac.New(sha256.New,[]byte(authSecret))
	sHash.Write(authChallenge)
	sig := sHash.Sum(nil)
	
	//Convert to ascii binary
	s := base64.StdEncoding.EncodeToString(sig)
	
	return s
}