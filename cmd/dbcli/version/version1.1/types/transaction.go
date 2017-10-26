package version1_1

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	//"github.com/hyperchain/hyperchain/crypto/guomi"
	//"github.com/hyperchain/hyperchain/admittance"
	"encoding/json"
	"github.com/hyperchain/hyperchain/crypto/sha3"
	"github.com/op/go-logging"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}
func (self *Transaction) BuildHash() common.Hash {
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

func (self *Transaction) GetTransactionHash() common.Hash {
	if len(self.TransactionHash) == 0 {
		return self.BuildHash()
	}
	return common.BytesToHash(self.TransactionHash)
}

func (self *Transaction) Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
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
	//sha3 Hash
	hashResult := ch.ByteHash([]byte(needHash))
	return hashResult
}

//func (self *Transaction) SignHashSM3(pubX,pubY []byte) []byte {
//	/*
//	from=0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd
//	&to=0x80958818f0a025273111fba92ed14c3dd483caeb
//	&value=0x08904e10904e1835
//	&timestamp=0x14a31c7e4883b166
//	&nonce=0x179a44e05e42f7
//	*/
//	h := guomi.New()
//	ENTL1 := "00"
//	h.Write(common.Hex2Bytes(ENTL1))
//	ENTL2 := "80"
//	h.Write(common.Hex2Bytes(ENTL2))
//	userId := "31323334353637383132333435363738"
//	h.Write(common.Hex2Bytes(userId))
//	a:= "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC"
//	h.Write(common.Hex2Bytes(a))
//	b:= "28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93"
//	h.Write(common.Hex2Bytes(b))
//	xG := "32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7"
//	h.Write(common.Hex2Bytes(xG))
//	yG := "BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0"
//	h.Write(common.Hex2Bytes(yG))
//	h.Write(pubX)
//	h.Write(pubY)
//	res := h.Sum(nil)
//
//	h2 := guomi.New()
//
//	value := new(TransactionValue)
//	hashErr := proto.Unmarshal(self.Value,value)
//	if hashErr != nil{
//		log.Error("cannot unmarshal the transaction value!")
//		h2.Write([]byte("invalid hash"))
//		return h2.Sum(nil)
//	}
//
//	var needHash string
//	if value.Payload == nil{
//		needHash = "from="+common.ToHex(self.From)+"&to="+common.ToHex(self.To)+"&value=0x"+strconv.FormatInt(value.Amount,16)+"&timestamp=0x"+strconv.FormatInt(self.Timestamp,16)+"&nonce=0x"+strconv.FormatInt(self.Nonce,16)
//	}else{
//		log.Debug("x: ",common.ToHex(value.Payload))
//		needHash = "from="+common.ToHex(self.From)+"&to="+common.ToHex(self.To)+"&value="+common.ToHex(value.Payload)+"&timestamp=0x"+strconv.FormatInt(self.Timestamp,16)+"&nonce=0x"+strconv.FormatInt(self.Nonce,16)
//	}
//	log.Debug(needHash)
//	//修改为sm3hash方法
//	h2.Write(res)
//	h2.Write([]byte(needHash))
//	hashResult := h2.Sum(nil)
//	return hashResult
//}

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
//func (self *Transaction) ValidateSign(encryption crypto.Encryption, ch crypto.CommonHash, ca *admittance.CAManager) bool {
//	if ca.GmSSL==true {
//		//sm2p256v1 := guomi.Curve(1)
//		//puk, err := guomi.ParsePublicKeyByDerEncode(sm2p256v1,self.Puk)
//		if len(self.Puk) > 65 {
//			log.Error("The Public Key is wrong!Publick Length is ",len(self.Puk),"!")
//			return false
//		}
//
//		//验证来源
//		var addr common.Address
//		copy(addr[:], Keccak256(self.Puk[0:])[12:])
//		from := common.BytesToAddress(self.From)
//		if from != addr{
//			log.Debug("From :",from.Hex())
//			log.Debug("Address :",addr.Hex())
//			log.Error("From address is wrong , please check the it!")
//			return false
//		}
//
//		puk, err := guomi.ParsePublicKeyByEncode(self.Puk)
//
//		if err != nil {
//			log.Error(err)
//			return false
//		}
//		hash := self.SignHashSM3(puk.X,puk.Y)
//
//		bol,err:= puk.VerifySignature(self.Signature,hash)
//		if err != nil {
//			log.Error(err)
//			return false
//		}
//		return bol
//	}
//	hash := self.SignHash(ch)
//	addr, err := encryption.UnSign(hash[:], self.Signature)
//	if err!=nil{
//		log.Error(err)
//		return false
//	}
//	from := common.BytesToAddress(self.From)
//	return addr==from
//
//}
// NewTransaction returns a new transaction
//func NewTransaction(from []byte,to []byte,value []byte, signature []byte) *Transaction{
func NewTransaction(from []byte, to []byte, value []byte, timestamp int64, nonce int64) *Transaction {

	transaction := &Transaction{
		From:  from,
		To:    to,
		Value: value,
		//Timestamp: time.Now().UnixNano(),
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

func ReadSourceFromFile(filePath string) string {

	fi, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	file_raw_source, err := ioutil.ReadAll(fi)
	defer fi.Close()
	return string(file_raw_source)
}

func NewTestCreateTransaction() *Transaction {
	// it is the code of hyperchain/core/vm/tests/solidity_files/example3.solc
	/*_,bins,_,err := compiler.CompileSourcefile(ReadSourceFromFile("github.com/hyperchain/hyperchain/core/vm/tests/solidity_files/test.solc"))
	if err!=nil{
		log.Errorf("the compiled source has error")
		return nil
	}*/

	//var code = bins[0]
	// it is the code of hyperchain/core/vm/tests/solidity_files/example1.solc
	var code = "0x60606040526000805463ffffffff191681557f6162636465666768696a6b6c6d6e6f707172737475767778797a00000000000060015560cf90819061004390396000f3606060405260e060020a60003504633ad14af38114603a578063569c5f6d1460615780638da9b772146074578063d09de08a146081575b6002565b346002576000805463ffffffff8116600435016024350163ffffffff19919091161790555b005b3460025760a360005463ffffffff165b90565b3460025760bd6001546071565b34600257605f6000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60408051918252519081900360200190f3"
	//var code = "0x60606040526000805463ffffffff191681557f6162636465666768696a6b6c6d6e6f707172737475767778797a00000000000060015560be90819061004390396000f3606060405260e060020a60003504633ad14af381146038578063569c5f6d14605c5780638da9b77214606b578063d09de08a146074575b005b6000805463ffffffff8116600435016024350163ffffffff19919091161790556036565b609260005463ffffffff165b90565b60ac6001546068565b60366000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60408051918252519081900360200190f3"
	var tx_value1 = &TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: common.FromHex(code)}
	value, _ := proto.Marshal(tx_value1)
	transaction := &Transaction{
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3").Bytes(),
		To:        nil,
		Value:     value,
		Timestamp: time.Now().UnixNano(),
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
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3").Bytes(),
		To:        nil,
		Value:     value,
		Timestamp: time.Now().UnixNano(),
	}

	return transaction
}

func NewTestCallTransaction() *Transaction {
	// it is the input of function add(uint32 num1,uint32 num2)
	var input = common.FromHex("0x8da9b772")
	//var input = common.FromHex("0x962b8398")

	// it is the input of function increment()
	//var input = common.FromHex("0xd09de08a")

	// it is the input of function getHello()
	//var input = common.FromHex("0x8da9b772")

	var tx_value2 = &TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: input}
	value, _ := proto.Marshal(tx_value2)
	transaction := &Transaction{
		From:      common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3").Bytes(),
		To:        common.HexToAddress("0x945304eb96065b2a98b57a48a06ae28d285a71b5").Bytes(),
		Value:     value,
		Timestamp: time.Now().UnixNano(),
	}

	return transaction
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
		Puk:             common.Bytes2Hex(self.Puk),
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
