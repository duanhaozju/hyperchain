//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"testing"
	"math/big"
	"fmt"
	"crypto/elliptic"
	"hyperchain/crypto/secp256k1"
	"sync/atomic"

	"crypto/ecdsa"
	"hyperchain/common"
)
type Transaction struct {
	data txdata
	// caches
	from atomic.Value
}
type txdata struct  {
	Recipient *common.Address
	Amount *big.Int
	signature []byte
}
func NewTransaction(to common.Address,amount *big.Int) *Transaction {
	d:=txdata{
		Recipient:	&to,
		Amount:		new(big.Int),
	}
	if amount != nil{
		d.Amount.Set(amount)
	}
	return &Transaction{data:d}
}

func TestSigntx(t *testing.T)  {
	ee := NewEcdsaEncrypto("ECDSAEncryto")
	k, err:= ee.GeneralKey()
	if err!=nil{
		panic(err)
	}

	key := k.(*ecdsa.PrivateKey)
	pub := key.PublicKey
	var addr common.Address
	pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
	copy(addr[:],Keccak256(pubBytes[1:])[12:])

	fmt.Println("public key is :")
	fmt.Println(pub)
	fmt.Println("private key is :")
	fmt.Println(key)
	//SaveNodeInfo("./port_address_privatekey","5004",addr,key)

	//p,err:=ee.GetKey()
	if err!=nil{
		panic(err)
	}
	priv := k.(*ecdsa.PrivateKey)

	//签名交易
	tx:= NewTransaction(common.Address{},big.NewInt(100))
	s256 := NewKeccak256Hash("Keccak256")
	hash := s256.Hash([]interface{}{tx.data.Amount,tx.data.Recipient})
	signature,err := ee.Sign(hash[:],priv)

	fmt.Println(signature)

	if err != nil {
	 	t.Error(err)
		t.FailNow()

	}
	//验证签名
	//from,err:= ee.UnSign(hash[:],signature)
	//if err != nil {
	//	t.Error(err)
	//	t.FailNow()
	//}
	//
	//fmt.Println(from)
	//fmt.Println(addr)
	//
	//hex := common.ToHex(from.Bytes())
	//fmt.Println(common.ToHex(from[:]))
	//fmt.Println(common.ToHex(addr[:]))
	//fmt.Println(common.FromHex(hex))

}
