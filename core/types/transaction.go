// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/crypto/guomi"
	"github.com/hyperchain/hyperchain/crypto/sha3"
	"github.com/op/go-logging"
	"math/big"
	"strconv"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("transaction")
}

func (self *Transaction) Hash() common.Hash {
	ch := crypto.NewKeccak256Hash("keccak256")
	return ch.Hash([]interface{}{
		self.From,
		self.To,
		self.Value,
		self.Timestamp,
		self.Nonce,
		self.Signature,
	})
}

func (self *Transaction) GetHash() common.Hash {
	if len(self.TransactionHash) == 0 {
		return self.Hash()
	}
	return common.BytesToHash(self.TransactionHash)
}

func (self *Transaction) SignHash(ch crypto.CommonHash) common.Hash {
	value := new(TransactionValue)
	hashErr := proto.Unmarshal(self.Value, value)
	if hashErr != nil {
		fmt.Println("cannot unmarshal the transaction value!")
		return ch.ByteHash([]byte("invalid hash"))
	}
	var needHash string
	if value.Payload == nil {
		needHash = "from=" + common.ToHex(self.From) + "&to=" + common.ToHex(self.To) + "&value=0x" + strconv.FormatInt(value.Amount, 16) + "&timestamp=0x" + strconv.FormatInt(self.Timestamp, 16) + "&nonce=0x" + strconv.FormatInt(self.Nonce, 16)
	} else {
		needHash = "from=" + common.ToHex(self.From) + "&to=" + common.ToHex(self.To) + "&value=" + common.ToHex(value.Payload) + "&timestamp=0x" + strconv.FormatInt(self.Timestamp, 16) + "&nonce=0x" + strconv.FormatInt(self.Nonce, 16)
	}
	hashResult := ch.ByteHash([]byte(needHash))
	return hashResult
}

func (self *Transaction) SignHashSM3(pubX, pubY []byte) []byte {
	h := guomi.New()
	ENTL1 := "00"
	h.Write(common.Hex2Bytes(ENTL1))
	ENTL2 := "80"
	h.Write(common.Hex2Bytes(ENTL2))
	userId := "31323334353637383132333435363738"
	h.Write(common.Hex2Bytes(userId))
	a := "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC"
	h.Write(common.Hex2Bytes(a))
	b := "28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93"
	h.Write(common.Hex2Bytes(b))
	xG := "32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7"
	h.Write(common.Hex2Bytes(xG))
	yG := "BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0"
	h.Write(common.Hex2Bytes(yG))
	h.Write(pubX)
	h.Write(pubY)
	res := h.Sum(nil)

	h2 := guomi.New()

	value := new(TransactionValue)
	hashErr := proto.Unmarshal(self.Value, value)
	if hashErr != nil {
		log.Error("cannot unmarshal the transaction value!")
		h2.Write([]byte("invalid hash"))
		return h2.Sum(nil)
	}

	var needHash string
	if value.Payload == nil {
		needHash = "from=" + common.ToHex(self.From) + "&to=" + common.ToHex(self.To) + "&value=0x" + strconv.FormatInt(value.Amount, 16) + "&timestamp=0x" + strconv.FormatInt(self.Timestamp, 16) + "&nonce=0x" + strconv.FormatInt(self.Nonce, 16)
	} else {
		needHash = "from=" + common.ToHex(self.From) + "&to=" + common.ToHex(self.To) + "&value=" + common.ToHex(value.Payload) + "&timestamp=0x" + strconv.FormatInt(self.Timestamp, 16) + "&nonce=0x" + strconv.FormatInt(self.Nonce, 16)
		needHash = "from=" + common.ToHex(self.From) + "&to=" + common.ToHex(self.To) + "&value=" + common.ToHex(value.Payload) + "&timestamp=0x" + strconv.FormatInt(self.Timestamp, 16) + "&nonce=0x" + strconv.FormatInt(self.Nonce, 16)
	}
	h2.Write(res)
	h2.Write([]byte(needHash))
	hashResult := h2.Sum(nil)
	return hashResult
}

func (self *Transaction) FString() string {
	return fmt.Sprintf(`
	From : %s
	To : %s
	Value : %s
	TimeStamp : %d
	Signature : %s`,
		self.From,
		self.To,
		self.Value,
		self.Timestamp,
		self.Signature)
}

//validates the signature
//if addr recovered from signature != tx.from return false
func (self *Transaction) ValidateSign(encryption crypto.Encryption, ch crypto.CommonHash) bool {
	if len(self.Signature) < 66 {
		log.Error("Illegal Signature length,please check it!")
		return false
	}
	flag := self.Signature[0]
	if flag == 1 {
		pub := make([]byte, 65)
		copy(pub[:], self.Signature[1:66])
		sign := make([]byte, len(self.Signature)-66)
		copy(sign[:], self.Signature[66:])

		var addr common.Address
		copy(addr[:], Keccak256(pub[0:])[12:])
		from := common.BytesToAddress(self.From)
		if from != addr {
			log.Error("From address is wrong , please check it!")
			return false
		}

		puk, err := guomi.ParsePublicKeyByEncode(pub)

		if err != nil {
			log.Error(err)
			return false
		}
		hash := self.SignHashSM3(puk.X, puk.Y)
		bol, err := puk.VerifySignature(sign, hash)
		if err != nil {
			log.Error(err)
			return false
		}
		return bol
	}
	hash := self.SignHash(ch)
	addr, err := encryption.UnSign(hash[:], self.Signature[1:])
	if err != nil {
		log.Error(err)
		return false
	}
	from := common.BytesToAddress(self.From)
	return addr == from

}

// NewTransaction returns a new transaction
//func NewTransaction(from []byte,to []byte,value []byte, signature []byte) *Transaction{
func NewTransaction(from []byte, to []byte, value []byte, timestamp int64, nonce int64) *Transaction {
	transaction := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		Timestamp: timestamp,
		Nonce:     nonce,
		Other:     &NonHash{},
	}

	return transaction
}

func NewTransactionValue(price, gasLimit, amount int64, payload []byte, opcode int32, extra []byte, vmType TransactionValue_VmType) *TransactionValue {
	return &TransactionValue{
		Price:    price,
		GasLimit: gasLimit,
		Amount:   amount,
		Payload:  payload,
		Extra:    extra,
		Op:       TransactionValue_Opcode(opcode),
		VmType:   vmType,
	}
}

func (tx *Transaction) GetTransactionValue() *TransactionValue {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return transactionValue
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func (tv *TransactionValue) RetrievePayload() []byte {
	return common.CopyBytes(tv.Payload)
}

func (tv *TransactionValue) RetrieveGas() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.GasLimit))
}

func (tv *TransactionValue) RetrieveGasPrice() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Price))
}

func (tv *TransactionValue) RetrieveAmount() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Amount))
}

func (tx *Transaction) GetNVPHash() string {
	return common.Bytes2Hex(tx.GetOther().NodeHash)
}

func (tx *Transaction) SetNVPHash(hash string) error {
	tx.GetOther().NodeHash = common.Hex2Bytes(hash)
	return nil
}
