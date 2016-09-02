/**
 * Created by Meiling Hu on 9/2/16.
 */
package accounts

import (
	"testing"
	"fmt"
	"encoding/hex"
	"hyperchain/crypto"
	"hyperchain/common"
)

func TestManager(t *testing.T)  {
	scryptN := StandardScryptN
	scryptP := StandardScryptP

	keydir := "/tmp/hyperchain/cache/keystore/"

	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := NewManager(keydir,encryption, scryptN, scryptP)
	account,err := am.NewAccount("123")
	if err!=nil{
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("------new account------")
	fmt.Println(account.Address)
	fmt.Println(common.ToHex(account.Address))

	fmt.Println("------get key according to the given account------")
	key,_ := am.GetDecryptedKey(account)
	fmt.Println(key.Address)
	fmt.Println(hex.EncodeToString(key.Address))
	fmt.Println(key.PrivateKey)

}