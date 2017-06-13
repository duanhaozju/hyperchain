// Package hts implements the Hyper Transport security
// include double side key agreement, and key session management
// this is the feature of hyperchian release 1.3 draft
// more details:
package hts

import (
	"crypto/rand"
	"fmt"
)



type HTS struct{
	crypter Crypto

}

//NegoShareKey use ecdh nego double side share key
func (hts *HTS)NegoShareKey(remotePubk, extra []byte)(shareKey []byte,err error){
	return hts.crypter.NegoSharedKey(remotePubk,extra)
}

//CSPRNG Cryptographically secure pseudorandom number generator
// CSPRNG generate the specific bit size Cryptoraphically random number
func CSPRNG(size int) (random []byte,err error){
	random  = make([]byte, size)
	_,err = rand.Read(random)
	if err != nil {
		fmt.Println("error:", err)
	}
	return
}


