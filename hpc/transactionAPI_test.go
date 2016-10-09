package hpc

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"os"
	"testing"
	"time"
)

var api = NewPublicTransactionAPI(nil, nil)
var from = common.HexToAddress("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
var to = common.HexToAddress("0x0000000000000000000000000000000000000003")
var args = SendTxArgs{
	From:     from,
	To:       &to,
	Gas:      NewInt64ToNumber(1000),
	GasPrice: NewInt64ToNumber(1000),
	Value:    NewInt64ToNumber(1000),
	Payload:  "",
}
var newTx common.Hash

func putTransactionToDefaultDB() {
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Error(err)
	}

	txValue := types.NewTransactionValue(100, 100, 1, nil)

	value, err := proto.Marshal(txValue)

	if err != nil {
		log.Error(err)
	}

	tx := types.NewTransaction(from[:], to[:], value)

	hash := tx.BuildHash()

	core.PutTransaction(db, hash[:], tx)
	newTx = hash
}

func TestPublicTransactionAPI_SendTransaction(t *testing.T) {

	if hash, err := api.SendTransaction(args); err == nil {
		t.Logf("the new tx hash is %v", common.ToHex(hash[:]))
	} else {
		t.Errorf("%v", err)
	}

}

func TestPublicTransactionAPI_SendTransactionOrContract(t *testing.T) {

	// deploy contract test
	args.To = nil
	args.Payload = "0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101cd806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003f578063569c5f6d146100675780638da9b7721461007c578063d09de08a146100ea575b610002565b34610002576000805460043563ffffffff8216016024350163ffffffff19919091161790555b005b346100025761010e60005463ffffffff165b90565b3461000257604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101289490928301828280156101c15780601f10610196576101008083540402835291602001916101c1565b34610002576100656000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b8154815290600101906020018083116101a457829003601f168201915b5050505050905061007956"

	if hash, err := api.SendTransactionOrContract(args); err == nil {
		t.Logf("the new contract tx hash is %v", common.ToHex(hash[:]))
	} else {
		t.Errorf("%v", err)
	}
}

func TestPublicTransactionAPI_GetTransactionByHash(t *testing.T) {

	putTransactionToDefaultDB()

	txGet, err := api.GetTransactionByHash(newTx)

	if err != nil {
		t.Errorf("%v", err)
	}

	t.Logf("get transaction by tx hash: %#v\n", common.ToHex(txGet.Hash[:]))
	e := json.NewEncoder(os.Stdout)
	e.Encode(txGet)
}

var txValue = types.NewTransactionValue(100, 100, 1, nil)

var value, err = proto.Marshal(txValue)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From:      from[:],
		To:        to[:],
		Value:     value,
		TimeStamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
	},
	&types.Transaction{
		From:      from[:],
		To:        to[:],
		Value:     value,
		TimeStamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature2"),
	},
	&types.Transaction{
		From:      from[:],
		To:        to[:],
		Value:     value,
		TimeStamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature3"),
	},
}

var blockUtilsCase = types.Block{
	ParentHash:   common.StringToHash("parenthash").Bytes(),
	BlockHash:    common.StringToHash("blockhash").Bytes(),
	Transactions: transactionCases,
	Timestamp:    time.Now().UnixNano(),
	MerkleRoot:   []byte("merkeleroot"),
	Number:       1,
	WriteTime:    time.Now().UnixNano() + int64(time.Second)/2,
}

func TestPublicTransactionAPI_GetTransactionByBlockHashAndIndex(t *testing.T) {

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = core.PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := api.GetTransactionByBlockHashAndIndex(common.BytesToHash(blockUtilsCase.BlockHash), 0)

	if err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("%#v", tx)
	}

}

func TestPublicTransactionAPI_GetTransactionsByBlockNumberAndIndex(t *testing.T) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = core.PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := api.GetTransactionsByBlockNumberAndIndex(1, 0)

	if err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("%#v", tx)
	}
}
