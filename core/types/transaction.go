package types

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/crypto"
	"strconv"
)

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

func (self *Transaction) SighHash(ch crypto.CommonHash) common.Hash {
	/*
	from=0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd
	&to=0x80958818f0a025273111fba92ed14c3dd483caeb
	&value=0x08904e10904e1835
	&timestamp=0x14a31c7e4883b166
	&nonce=0x179a44e05e42f7
	*/
	value := new(TransactionValue)
	hashErr := proto.Unmarshal(self.Value,value)
	if hashErr != nil{
		fmt.Println("cannot unmarshal the transaction value!")
		return ch.ByteHash([]byte("invalid hash"))
	}
	var needHash string
	if value.Payload == nil{
		needHash = "from="+common.ToHex(self.From)+"&to="+common.ToHex(self.To)+"&value=0x"+strconv.FormatInt(value.Amount,16)+"&timestamp=0x"+strconv.FormatInt(self.Timestamp,16)+"&nonce=0x"+strconv.FormatInt(self.Nonce,16)
	}else{
		needHash = "from="+common.ToHex(self.From)+"&to="+common.ToHex(self.To)+"&value="+common.ToHex(value.Payload)+"&timestamp=0x"+strconv.FormatInt(self.Timestamp,16)+"&nonce=0x"+strconv.FormatInt(self.Nonce,16)
	}
	hashResult := ch.ByteHash([]byte(needHash))
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

	hash := self.SighHash(ch)
	addr, _ := encryption.UnSign(hash[:], self.Signature)
	from := common.BytesToAddress(self.From)
	return addr == from
}

// NewTransaction returns a new transaction
//func NewTransaction(from []byte,to []byte,value []byte, signature []byte) *Transaction{
func NewTransaction(from []byte, to []byte, value []byte, timestamp int64, nonce int64) *Transaction {
	transaction := &Transaction{
		From:  from,
		To:    to,
		Value: value,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	return transaction
}

func NewTransactionValue(price, gasLimit, amount int64, payload []byte, opcode int32) *TransactionValue {
	return &TransactionValue{
		Price:    price,
		GasLimit: gasLimit,
		Amount:   amount,
		Payload:  payload,
		Op:       TransactionValue_Opcode(opcode),
	}
}

func (tx *Transaction) GetTransactionValue() *TransactionValue {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return transactionValue
}
