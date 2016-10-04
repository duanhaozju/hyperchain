package types

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/crypto"
	"math/big"
	"time"
)

func (self *Transaction) BuildHash() common.Hash {
	ch := crypto.NewKeccak256Hash("keccak256")
	return ch.Hash(self)
}

func (self *Transaction) Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}

func (self *Transaction) SighHash(ch crypto.CommonHash) common.Hash {
	return ch.Hash([]interface{}{
		self.Value,
		self.TimeStamp,
		self.From,
		self.To,
	})
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
		self.TimeStamp,
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
func NewTransaction(from []byte, to []byte, value []byte) *Transaction {

	transaction := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}

func NewTransactionValue(price, gasLimit, amount int64, payload []byte) *TransactionValue {
	return &TransactionValue{
		Price:    price,
		GasLimit: gasLimit,
		Amount:   amount,
		Payload:  payload,
	}
}

func NewTestCreateTransaction() *Transaction {
	// it is the code of hyperchain/core/vm/tests/solidity_files/example3.solc
	var code = "0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101b8806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003c578063569c5f6d146100615780638da9b77214610071578063d09de08a146100da575b005b6000805460043563ffffffff8216016024350163ffffffff199190911617905561003a565b6100f960005463ffffffff165b90565b604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101139490928301828280156101ac5780601f10610181576101008083540402835291602001916101ac565b61003a6000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101735780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b81548152906001019060200180831161018f57829003601f168201915b5050505050905061006e56"
	// it is the code of hyperchain/core/vm/tests/solidity_files/example1.solc
	//var code = "0x60606040526000805463ffffffff191681557f6162636465666768696a6b6c6d6e6f707172737475767778797a00000000000060015560be90819061004390396000f3606060405260e060020a60003504633ad14af381146038578063569c5f6d14605c5780638da9b77214606b578063d09de08a146074575b005b6000805463ffffffff8116600435016024350163ffffffff19919091161790556036565b609260005463ffffffff165b90565b60ac6001546068565b60366000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60408051918252519081900360200190f3"
	var tx_value1 = &TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: common.FromHex(code)}
	value, _ := proto.Marshal(tx_value1)
	transaction := &Transaction{
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To:        nil,
		Value:     value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}

func NewTestCreateTransactionSourceCode() *Transaction {
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
	var tx_value1 = &TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: ([]byte)(sourcecode)}
	value, _ := proto.Marshal(tx_value1)
	transaction := &Transaction{
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To:        nil,
		Value:     value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}

func NewTestCallTransaction() *Transaction {
	// it is the input of function add(uint32 num1,uint32 num2)
	var input = common.FromHex("0x8da9b772")

	// it is the input of function increment()
	//var input = common.FromHex("0xd09de08a")

	// it is the input of function getHello()
	//var input = common.FromHex("0x8da9b772")

	var tx_value2 = &TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: input}
	value, _ := proto.Marshal(tx_value2)
	transaction := &Transaction{
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To:        common.HexToAddress("0x945304eb96065b2a98b57a48a06ae28d285a71b5").Bytes(),
		Value:     value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}
func (tx *Transaction) Payload() []byte {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return common.CopyBytes(transactionValue.Payload)
}
func (tx *Transaction) Gas() *big.Int {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.GasLimit))
}
func (tx *Transaction) GasPrice() *big.Int {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.Price))
}
func (tx *Transaction) Amount() *big.Int {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value, transactionValue)
	return new(big.Int).Set(big.NewInt(transactionValue.Amount))
}
