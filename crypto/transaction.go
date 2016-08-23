// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (

	"crypto/ecdsa"
	"errors"
	"fmt"

	"math/big"

	"sync/atomic"




	//"hyperChain/rlp"
	"hyperchain-alpha/crypto/sha3"
	"hyperchain-alpha/common"
	"encoding/json"
)

var ErrInvalidSig = errors.New("invalid v, r, s values")

type Transaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	Recipient       *common.Address `rlp:"nil"` // nil means contract creation
	Amount          *big.Int
	V               byte     // signature
	R, S            *big.Int // signature
}



func NewTransaction(nonce uint64, to common.Address, amount, gasLimit, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		Recipient:    &to,
		Amount:       new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}

	return &Transaction{data: d}
}


func (tx *Transaction) Value() *big.Int    { return new(big.Int).Set(tx.data.Amount) }


//hash 方法
func rlpHash(x interface{}) (h common.Hash) {

	hw := sha3.NewKeccak256()
	data,err :=json.Marshal(x)
	hw.Write(data)
	if err!=nil{
		fmt.Print("2")
	}

	//rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}


func (tx *Transaction) SigHash() common.Hash {
	return rlpHash([]interface{}{

		tx.data.Recipient,
		tx.data.Amount,

	})
}
type writeCounter common.StorageSize
func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// 得到签名后的transaction的from,如果error,则解密失败
func (tx *Transaction) From() (common.Address, error) {
	return doFrom(tx, true)
}



func doFrom(tx *Transaction, homestead bool) (common.Address, error) {
	if from := tx.from.Load(); from != nil {
		return from.(common.Address), nil
	}
	pubkey, err := tx.publicKey(homestead)
	if err != nil {
		return common.Address{}, err
	}
	var addr common.Address
	copy(addr[:], Keccak256(pubkey[1:])[12:])
	tx.from.Store(addr)
	return addr, nil
}


//根据transaction中途的r s v获得该from账户的public key
func (tx *Transaction) publicKey(homestead bool) ([]byte, error) {
	if !ValidateSignatureValues(tx.data.V, tx.data.R, tx.data.S, homestead) {
		return nil, ErrInvalidSig
	}

	// encode the signature in uncompressed format
	r, s := tx.data.R.Bytes(), tx.data.S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = tx.data.V - 27

	// recover the public key from the signature
	hash := tx.SigHash()
	pub, err := Ecrecover(hash[:], sig)
	if err != nil {
		//glog.V(logger.Error).Infof("Could not get pubkey from signature: ", err)
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}
	return pub, nil
}

func (tx *Transaction) WithSignature(sig []byte) (*Transaction, error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R = new(big.Int).SetBytes(sig[:32])
	cpy.data.S = new(big.Int).SetBytes(sig[32:64])
	cpy.data.V = sig[64] + 27
	return cpy, nil
}

func (tx *Transaction) SignECDSA(prv *ecdsa.PrivateKey) (*Transaction, error) {
	h := tx.SigHash()
	sig, err := Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(sig)
}

/*func (tx *Transaction) String() string {
	var from, to string
	if f, err := tx.From(); err != nil {
		from = "[invalid sender]"
	} else {
		from = fmt.Sprintf("%x", f[:])
	}
	if tx.data.Recipient == nil {
		to = "[contract creation]"
	} else {
		to = fmt.Sprintf("%x", tx.data.Recipient[:])
	}
	enc, _ := rlp.EncodeToBytes(&tx.data)
	return fmt.Sprintf(`
	TX(%x)
	Contract: %v
	From:     %s
	To:       %s
	Nonce:    %v
	GasPrice: %v
	GasLimit  %v
	Value:    %v
	Data:     0x%x
	V:        0x%x
	R:        0x%x
	S:        0x%x
	Hex:      %x
`,
		tx.Hash(),
		len(tx.data.Recipient) == 0,
		from,
		to,
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.GasLimit,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.V,
		tx.data.R,
		tx.data.S,
		enc,
	)
}*/

// Transaction slice type for basic sorting.
type Transactions []*Transaction










