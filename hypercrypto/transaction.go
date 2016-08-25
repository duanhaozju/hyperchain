package hyperencrypt

import (
	"sync/atomic"
	"math/big"
	"crypto/ecdsa"
	"fmt"
	"errors"
	"hyperchain-alpha/common"
)
type Transaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
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
func (tx *Transaction) Hash() common.Hash{
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := Fix32Hash(tx)
	tx.hash.Store(v)
	return v
}
func (tx *Transaction)SigHash() common.Hash  {
	return Fix32Hash([]interface{}{
		tx.data.Recipient,
		tx.data.Amount,
	})
}
func (tx *Transaction) WithSignature(sig []byte) (*Transaction,error){
	if len(sig)!=65 {
		panic(fmt.Sprintf("wrong size for signature: got %s, want 65",len(sig)))
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.signature = sig
	return cpy,nil
}
func (tx *Transaction)SignECDSA(prv *ecdsa.PrivateKey) (*Transaction,error) {
	h := tx.SigHash()
	sig,err :=Sign(h[:],prv)
	if err!=nil{
		 return nil,err
	 }
	return tx.WithSignature(sig)
}
func (tx *Transaction) From() (common.Address,error){
	return doFrom(tx,true)
}
func  doFrom(tx *Transaction,homestead bool) (common.Address,error){
	if from := tx.from.Load();from!=nil{
		return from.(common.Address),nil
	}
	pubKey,err := tx.publicKey(homestead)
	if err!=nil{
		return common.Address{},err
	}
	var addr common.Address
	copy(addr[:],Keccak256(pubKey[1:])[12:])
	tx.from.Store(addr)
	return addr,nil

}
func (tx *Transaction)publicKey(homestead bool)([]byte,error)  {
	sig := make([]byte,65)
	copy(sig[:],tx.data.signature)
	hash := tx.SigHash()
	pub,err := Ecrecover(hash[:],sig)
	if err != nil {
		//glog.V(logger.Error).Infof("Could not get pubkey from signature: ", err)
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}
	return pub, nil
}