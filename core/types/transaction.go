package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"fmt"
	"math/big"
	"github.com/golang/protobuf/proto"
	"github.com/labstack/gommon/log"
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
	log.Info("=======enter NewTransactionValue ================")
	log.Info(price,gasLimit,amount,payload)
	return &TransactionValue{
		Price: price,
		GasLimit: gasLimit,
		Amount: amount,
		Payload: payload,
	}
}

func NewTestCreateTransaction() *Transaction{
	// it is the code of hyperchain/core/vm/tests/solidity_files/example3.solc
	var code = "0x60606040526040516102763803806102768339810160405280510160008054600160a060020a031916331790558060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1060a057805160ff19168380011785555b50608f9291505b8082111560cd5760008155600101607d565b5050506101a5806100d16000396000f35b828001600101855582156076579182015b82811115607657825182600050559160200191906001019060b1565b509056606060405260e060020a60003504633ad14af3811461003157806341c0e1b514610055578063cfae321714610097575b005b6002805463ffffffff8116600435016024350163ffffffff1990911617905561002f565b61002f6000543373ffffffffffffffffffffffffffffffffffffffff9081169116141561016e5760005473ffffffffffffffffffffffffffffffffffffffff16ff5b604080516020818101835260008252600180548451600282841615610100026000190190921691909104601f810184900484028201840190955284815261010094909283018282801561019b5780601f106101705761010080835404028352916020019161019b565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101605780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b565b820191906000526020600020905b81548152906001019060200180831161017e57829003601f168201915b505050505090509056"
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
	var input = common.FromHex("0x3ad14af300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002")

	// it is the input of function increment()
	//var input = common.FromHex("0xd09de08a")

	// it is the input of function getHello()
	//var input = common.FromHex("0x8da9b772")


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
