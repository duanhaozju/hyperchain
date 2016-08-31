/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"testing"
	"fmt"
	"hyperchain/common"
)

func TestManager(t *testing.T) {
	am := NewManager()
	key ,err :=am.NewAccount("123")
	if err != nil{
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("********************")
	fmt.Println(key.auth)
	fmt.Println(key.address)
	fmt.Println(common.ToHex(key.address))

	//kk,_ := am.GetAccountKey(key.address,key.auth)
	//fmt.Println(kk)
	//fmt.Println(kk.address)

}
