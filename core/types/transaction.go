package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"fmt"
	"math/big"
	"github.com/golang/protobuf/proto"
)

func (self *Transaction)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}

func (self *Transaction)SighHash(ch crypto.CommonHash) common.Hash {
	return ch.Hash([]interface{}{
		self.Value,
		self.TimeStamp,
		self.From,
		self.To,
	})
}

func (self *Transaction)FString() string {
	return fmt.Sprintf(`
	From : %s
	To : %s
	Value : %s
	TimeStamp : %d
	Signature : %s`,
		self.From,
		self.To,
		self.Value,
		self.TimeStamp,
		self.Signature)
}
//validates the signature
//if addr recovered from signature != tx.from return false
func (self *Transaction)ValidateSign(encryption crypto.Encryption,ch crypto.CommonHash)  bool{

	hash := self.SighHash(ch)
	addr,_ := encryption.UnSign(hash[:],self.Signature)
	from := common.HexToAddress(string(self.From))
	return addr==from
}

// NewTransaction returns a new transaction
func NewTransaction(from []byte,to []byte,value []byte) *Transaction{

	transaction := &Transaction{
		From: from,
		To: to,
		Value: value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}
func NewTestCreateTransaction() *Transaction{
	var sourcecode = `
contract mortal {
     /* Define variable owner of the type address*/
     address owner;

     /* this function is executed at initialization and sets the owner of the contract */
     function mortal() {
         owner = msg.sender;
     }

     /* Function to recover the funds on the contract */
     function kill() {
         if (msg.sender == owner)
             selfdestruct(owner);
     }
 }


 contract greeter is mortal {
     /* define variable greeting of the type string */
     string greeting;
    uint32 sum;
     /* this runs when the contract is executed */
     function greeter(string _greeting) public {
         greeting = _greeting;
     }

     /* main function */
     function greet() constant returns (string) {
         return greeting;
     }
    function add(uint32 num1,uint32 num2) {
        sum = sum+num1+num2;
    }
 }`
	var tx_value1 = &TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:([]byte)(sourcecode)}
	value,_ := proto.Marshal(tx_value1)
	transaction := &Transaction{
		From:	common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To: nil,
		Value: value	,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}

func NewTestCallTransaction() *Transaction{
	var input = common.FromHex("0x3ad14af300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002")
	var tx_value2 = &TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:input}
	value,_ := proto.Marshal(tx_value2)
	transaction := &Transaction{
		From:	common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To: common.HexToAddress("0x945304eb96065b2a98b57a48a06ae28d285a71b5").Bytes(),
		Value: 	value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}
func (tx *Transaction) Payload() []byte       {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value,transactionValue)
	return common.CopyBytes(transactionValue.Payload)
}
func (tx *Transaction) Gas() *big.Int      {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value,transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.GasLimit))
}
func (tx *Transaction) GasPrice() *big.Int {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value,transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.Price))
}
func (tx *Transaction) Amount() *big.Int    {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value,transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.Amount))
}
