// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

package crypto

import (
	"testing"

	"fmt"
)


func TestEncryption(t *testing.T){
	encryption :=NewEcdsaEncrypto("ecdsa")
	encryption.GeneralKey("123")
	encryption.GetKey()
	fmt.Print(encryption.port)



}
