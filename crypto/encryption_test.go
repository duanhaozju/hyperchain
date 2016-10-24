// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

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
