// fetcher implements block operate
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-08-29
package core

import (
	"encoding/json"
	"hyperchain/core/types"
	"io/ioutil"

	"hyperchain/common"

	"hyperchain/hyperdb"

	"hyperchain/core/state"

	"math/big"
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

	var genesis = map[string]Genesis{}

	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Error("ReadFile: ", err.Error())
		return
	}

	if err := json.Unmarshal(bytes, &genesis); err != nil {
		log.Error("Unmarshal: ", err.Error())
		return
	}

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}

	stateDB, _ := state.New(common.Hash{}, db)
	for addr, account := range genesis["test1"].Alloc {
		/*balance:=types.Balance{
			AccountPublicKeyHash:[]byte(addr),
			Value:account,
		}*/
		//address := common.HexToAddress(addr)

		//statedb.AddBalance(address, common.String2Big(account))
		object := stateDB.CreateAccount(common.HexToAddress(addr))

		object.AddBalance(big.NewInt(account))

		/*
			balanceIns.PutCacheBalance(common.HexToAddress(addr),[]byte(account))
			balanceIns.PutDBBalance(common.HexToAddress(addr),[]byte(account))*/

	}
	stateDB.Commit()
	//stateDB.GetBalance()

	block := types.Block{
		ParentHash: common.FromHex(genesis["test1"].ParentHash),
		Timestamp:  genesis["test1"].Timestamp,
		BlockHash:  common.FromHex(genesis["test1"].BlockHash),
		Number:     genesis["test1"].Number,
		//MerkleRoot:       "root",
	}

	log.Debug("构造创世区块")
	err = PutBlock(db, block.BlockHash, &block)
	// write transaction
	//PutTransactions(db, commonHash, block.Transactions)
	if err != nil {
		log.Fatal(err)
	}
	UpdateChain(&block, true)

	log.Info("current chain block number is", GetChainCopy().Height)

}
