//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"hyperchain/core/types"
	"io/ioutil"

	"hyperchain/common"

	"hyperchain/hyperdb"

	//"fmt"
	"github.com/buger/jsonparser"
	"hyperchain/core/state"
	"math/big"
	"strconv"
	"time"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"errors"
	"hyperchain/tree/bucket"
)

func CreateInitBlock(filename string, stateType string, blockVersion string, bktConf bucket.Conf) {
	log.Info("genesis start")

	if GetHeightOfChain() > 0 {
		log.Info("already genesis")
		return
	}
	type Genesis struct {
		Timestamp  int64
		ParentHash string
		BlockHash  string
		Coinbase   string
		Number     uint64
		Alloc      map[string]int64
	}


	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Error("ReadFile: ", err.Error())
		return
	}

	// start the parse genesis content
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
		return
	}
	stateDB, err := GetStateInstance(common.Hash{}, db, stateType, bktConf)
	stateDB.MarkProcessStart(0)
	if err != nil {
		log.Error("genesis create statedb failed!")
		return
	}

	jsonparser.ObjectEach(bytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		object := stateDB.CreateAccount(common.HexToAddress(string(key)))
		account, _ := strconv.ParseInt(string(value), 10, 64)
		object.AddBalance(big.NewInt(account))
		return nil
	}, "genesis", "alloc")
	root, err := stateDB.Commit()

	if err != nil {
		log.Error("Genesis.go file statedb commit failed!")
		return
	}
	// flush state change to disk immediately
	batch := stateDB.FetchBatch(0)
	if batch != nil {
		batch.Write()
	}

	block := types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Timestamp:  time.Now().Unix(),
		BlockHash:  common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Number:     uint64(0),
		MerkleRoot: root.Bytes(),
	}

	log.Debug("construct genesis block")
	// flush block content to disk immediately
	if err, _ := PersistBlock(db.NewBatch(), &block, blockVersion, true, true); err != nil {
		log.Fatal(err)
		return
	}
	// flush change of chain to disk immediately
	UpdateChain(db.NewBatch(), &block, true, true, true)
	stateDB.MarkProcessFinish(0)
	log.Criticalf("Genesis state %s", string(stateDB.Dump(0)))
	log.Info("current chain block number is", GetChainCopy().Height)

}
func GetStateInstance(root common.Hash, db hyperdb.Database, stateType string, bktConf bucket.Conf) (vm.Database, error) {
	switch stateType {
	case "rawstate":
		return state.New(root, db)
	case "hyperstate":
		return hyperstate.New(root, db, bktConf, 0)
	default:
		return nil, errors.New("no state type specified")
	}
}
