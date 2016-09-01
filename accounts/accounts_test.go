/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"testing"
	"fmt"
	"math/big"
	"sync/atomic"
	"hyperchain/crypto"
	"hyperchain/common"
)
type Transaction struct {
	data txdata
	// caches
	from atomic.Value
}
type txdata struct  {
	Recipient *[]byte
	Amount *big.Int
	signature []byte
}
func NewTransaction(to []byte,amount *big.Int) *Transaction {
	d:=txdata{
		Recipient:	&to,
		Amount:		new(big.Int),
	}
	if amount != nil{
		d.Amount.Set(amount)
	}
	return &Transaction{data:d}
}

func TestManager(t *testing.T) {
	am := NewManager()
	key ,err :=am.NewAccount("123")
	if err != nil{
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("******newaccount*******")
	fmt.Println(key.Address)
	fmt.Println(key.Auth)
	fmt.Println(key.PrivateKey)
	kk, err:= am.GetAccountKey(key.Address,key.Auth)
	if err != nil{
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("******getkeyreturn*******")
	fmt.Println(kk.Address)
	fmt.Println(kk.Auth)
	fmt.Println(kk.PrivateKey)

	tx:= NewTransaction([]byte{},big.NewInt(100))
	s256 := crypto.NewKeccak256Hash("Keccak256")
	hash := s256.Hash([]interface{}{tx.data.Amount,tx.data.Recipient})

	//签名交易
	signature,err := am.encryption.Sign(hash[:],kk.PrivateKey)

	if err != nil {
		t.Error(err)
		t.FailNow()

	}
	//验证签名
	from,err:= am.encryption.UnSign(hash[:],signature)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(from)
	fmt.Println(common.ToHex(from))

}
