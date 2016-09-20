package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"fmt"
	"math/big"
	"github.com/golang/protobuf/proto"
	"hyperchain/core"
)

func (self *Transaction)BuildHash() common.Hash {
	ch := crypto.NewKeccak256Hash("keccak256")
	return ch.Hash(self)
}

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


func NewTransactionValue(price,gasLimit, amount int64, payload []byte) *TransactionValue{
	return &TransactionValue{
		Price: price,
		GasLimit: gasLimit,
		Amount: amount,
		Payload: payload,
	}
}

func NewTestCreateTransaction() *Transaction{
	// it is the code of hyperchain/core/vm/tests/solidity_files/example3.solc
	var code = "60606040526000600060006101000a81548163ffffffff02191690830217905550604060405190810160405280600981526020017f616161626262636363000000000000000000000000000000000000000000000081526020015060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100a857805160ff19168380011785556100d9565b828001600101855582156100d9579182015b828111156100d85782518260005055916020019190600101906100ba565b5b50905061010491906100e6565b8082111561010057600081815060009055506001016100e6565b5090565b5050610272806101146000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480633ad14af31461005a578063569c5f6d1461007b5780638da9b772146100a4578063d09de08a1461011f57610058565b005b610079600480803590602001909190803590602001909190505061012e565b005b6100886004805050610164565b604051808263ffffffff16815260200191505060405180910390f35b6100b16004805050610183565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101115780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b61012c600480505061023f565b005b8082600060009054906101000a900463ffffffff160101600060006101000a81548163ffffffff021916908302179055505b5050565b6000600060009054906101000a900463ffffffff169050610180565b90565b602060405190810160405280600081526020015060016000508054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156102305780601f1061020557610100808354040283529160200191610230565b820191906000526020600020905b81548152906001019060200180831161021357829003601f168201915b5050505050905061023c565b90565b6001600060009054906101000a900463ffffffff1601600060006101000a81548163ffffffff021916908302179055505b56"
	// it is the code of hyperchain/core/vm/tests/solidity_files/example1.solc
	//var code = "0x60606040526000805463ffffffff191681557f6162636465666768696a6b6c6d6e6f707172737475767778797a00000000000060015560be90819061004390396000f3606060405260e060020a60003504633ad14af381146038578063569c5f6d14605c5780638da9b77214606b578063d09de08a146074575b005b6000805463ffffffff8116600435016024350163ffffffff19919091161790556036565b609260005463ffffffff165b90565b60ac6001546068565b60366000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60408051918252519081900360200190f3"
	var tx_value1 = &TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:common.FromHex(code)}
	value,_ := proto.Marshal(tx_value1)
	transaction := &Transaction{
		From:	common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To: nil,
		Value: value	,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}


func NewTestCreateTransactionSourceCode() *Transaction{
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
	// it is the input of function add(uint32 num1,uint32 num2)
	var input = common.FromHex("0x8da9b772")

	// it is the input of function increment()
	//var input = common.FromHex("0xd09de08a")

	// it is the input of function getHello()
	//var input = common.FromHex("0x8da9b772")


	var tx_value2 = &TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:input}
	value,_ := proto.Marshal(tx_value2)
	transaction := &Transaction{
		From:	common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6").Bytes(),
		To: core.GetVMEnv().State().GetLeastAccount().Address().Bytes(),
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
