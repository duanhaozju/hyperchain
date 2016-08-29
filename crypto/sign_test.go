package crypto

import (
	"testing"
	"math/big"
	"fmt"
	"crypto/elliptic"
	"hyperchain/common"
	"hyperchain/crypto/secp256k1"
	"sync/atomic"
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
	key, _:= ee.GenerateKey()
	ee.SaveECDSA("./testFile",key)
	priv,_:=ee.LoadECDSA("./testFile")
	pub := key.PublicKey

	var addr common.Address
	pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
	addr = common.BytesToAddress(ee.Keccak256(pubBytes[1:])[12:])

	fmt.Println("public key is :")
	fmt.Println(pub)
	fmt.Println("private key is :")
	fmt.Println(key)

	//签名交易
	tx:= NewTransaction(common.Address{},big.NewInt(100))
	s256 := NewKeccak256Hash("Keccak256")
	hash := s256.Hash([]interface{}{tx.data.Amount,tx.data.Recipient})
	signature,err := ee.Sign(hash[:],*priv)
	if err != nil {
		t.Error(err)
		t.FailNow()

	}
	//验证签名
	from,err:= ee.UnSign(hash[:],signature)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(from)
	fmt.Println(addr)
	ee.SaveNodeInfo("./addressInfo","0.0.0.0",addr,key)

}
