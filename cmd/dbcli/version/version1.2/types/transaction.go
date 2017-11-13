package version1_2

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto"
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

func (self *Transaction) SignHash(ch crypto.CommonHash) common.Hash {
	/*
		from=0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd
		&to=0x80958818f0a025273111fba92ed14c3dd483caeb
		&value=0x08904e10904e1835
		&timestamp=0x14a31c7e4883b166
		&nonce=0x179a44e05e42f7
	*/
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

	hash := self.SignHash(ch)
	addr, _ := encryption.UnSign(hash[:], self.Signature)
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

func (self *Transaction) ToTransactionView() *TransactionView {
	transactionView := &TransactionView{
		Version:         string(self.Version),
		From:            common.Bytes2Hex(self.From),
		To:              common.Bytes2Hex(self.To),
		Value:           common.Bytes2Hex(self.Value),
		Timestamp:       self.Timestamp,
		Signature:       common.Bytes2Hex(self.Signature),
		Id:              self.Id,
		TransactionHash: common.Bytes2Hex(self.TransactionHash),
		Nonce:           self.Nonce,
	}
	return transactionView
}

func (self *Transaction) ToTransactionViewHash() *TransactionViewHash {
	transactionViewHash := &TransactionViewHash{
		TransactionHash: common.Bytes2Hex(self.TransactionHash),
	}
	return transactionViewHash
}

func (self *InvalidTransactionRecord) Encode() string {
	transactionView := self.Tx.ToTransactionView()
	invalidTransactionView := &InvalidTransactionView{
		Tx:      transactionView,
		ErrType: self.encodeErrType(self.ErrType),
		ErrMsg:  string(self.ErrMsg),
	}
	res, err := json.MarshalIndent(invalidTransactionView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (self *TransactionMeta) Encode() string {
	res, err := json.MarshalIndent(self, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (self *InvalidTransactionRecord) encodeErrType(errType InvalidTransactionRecord_ErrType) string {
	switch errType {
	case InvalidTransactionRecord_OUTOFBALANCE:
		return "OUTOFBALANCE"
	case InvalidTransactionRecord_SIGFAILED:
		return "SIGFAILED"
	case InvalidTransactionRecord_INVOKE_CONTRACT_FAILED:
		return "INVOKE_CONTRACT_FAILED"
	case InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED:
		return "DEPLOY_CONTRACT_FAILED"
	case InvalidTransactionRecord_INVALID_PERMISSION:
		return "INVALID_PERMISSION"
	default:
		return ""
	}
}
