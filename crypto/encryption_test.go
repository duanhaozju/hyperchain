//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"testing"

)


func TestEncryption(t *testing.T){
	encryption :=NewEcdsaEncrypto("ecdsa")
	//encryption.GeneralKey("123")
	//encryption.GetKey()
	encryption.GenerateNodeKey("123","../keystore/")
	//key ,_ := encryption.GetNodeKey("../keystore/")



}
