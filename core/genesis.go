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
	"hyperchain/crypto"
	"math/big"
	"strconv"
	"time"
)

func CreateInitBlock(filename string) {
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

	//var genesis = map[string]Genesis{}

	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Error("ReadFile: ", err.Error())
		return
	}

	// start  the parse genesis content

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
		return
	}

	stateDB, _ := state.New(common.Hash{}, db)

	// You can use `ObjectEach` helper to iterate objects { "key1":object1, "key2":object2, .... "keyN":objectN }
	jsonparser.ObjectEach(bytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		//fmt.Printf("Key: '%s'\n Value: '%s'\n Type: %s\n", string(key), string(value), dataType)
		object := stateDB.CreateAccount(common.HexToAddress(string(key)))
		account, _ := strconv.ParseInt(string(value), 10, 64)
		object.AddBalance(big.NewInt(account))
		return nil
	}, "genesis", "alloc")

	root, _ := stateDB.Commit()

	block := types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Timestamp:  time.Now().Unix(),
		BlockHash:  common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Number:     uint64(0),
		MerkleRoot: root.Bytes(),
	}

	log.Debug("构造创世区块")
	//if err := PutBlock(db, block.BlockHash, &block); err != nil {
	commonHash := crypto.NewKeccak256Hash("keccak256")
	if err := PutBlockTx(db, commonHash, block.BlockHash, &block); err != nil {
		log.Fatal(err)
		return
	}
	UpdateChain(&block, true)
	log.Info("current chain block number is", GetChainCopy().Height)

}
