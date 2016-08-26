package crypto

import (
	"sync/atomic"
	"math/big"
	"fmt"
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
	s256 := NewKeccak256Hash("Keccak256")
	v := s256.Hash(tx)
	tx.hash.Store(v)
	return v
}
func (tx *Transaction)SigHash() common.Hash  {
	s256 := NewKeccak256Hash("Keccak256")
	return s256.Hash([]interface{}{
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
