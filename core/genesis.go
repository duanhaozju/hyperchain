// fetcher implements block operate
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-08-29
package core

import (
	"fmt"
	"hyperchain/core/types"
	"io/ioutil"
	"encoding/json"

	"hyperchain/common"

	"log"

	"hyperchain/hyperdb"

	"hyperchain/crypto"
)


func CreateInitBlock(filename string)  {
	log.Println("genesis start")

	type Genesis struct {
		Timestamp  int64
		ParentHash string
		BlockHash  string
		Coinbase   string
		Number     uint64
		Alloc      map[string]string
	}

	var genesis = map[string]Genesis{}

	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
		return
	}

	if err := json.Unmarshal(bytes, &genesis); err != nil {
		fmt.Println("Unmarshal: ", err.Error())
		return
	}
	
	balanceIns, err := GetBalanceIns()
	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}
	for addr, account := range genesis["test1"].Alloc {
		//address := common.HexToAddress(addr)

		//value, err := strconv.ParseInt(account.Balance, 10, 64)
		//fmt.Println(addr)
		//fmt.Println([]byte(addr))
		//fmt.Println(common.BytesToHash([]byte(addr)))
		//fmt.Println(common.BytesToAddress([]byte("0000000000000000000000000000000000000002")))
		/*balance:=types.Balance{
			AccountPublicKeyHash:[]byte(addr),
			Value:account,
		}*/

		balanceIns.PutCacheBalance(common.BytesToAddress([]byte(addr)),[]byte(account))
		balanceIns.PutDBBalance(common.BytesToAddress([]byte(addr)),[]byte(account))


	}
	db,err:=hyperdb.GetLDBDatabase()
	PutDBBalance(db,balanceIns.dbBalance)
	if err!=nil{
		log.Fatal(err)
	}


	block := types.Block{
		ParentHash: common.FromHex(genesis["test1"].ParentHash),
		Timestamp:   genesis["test1"].Timestamp,
		BlockHash: common.FromHex(genesis["test1"].BlockHash),
		Number:   genesis["test1"].Number,
		//MerkleRoot:       "root",
	}



	log.Println("构造创世区块")

	UpdateChain(block.BlockHash,true)
	log.Println("current chain block number is",GetChainCopy().Height)

}

// WriteBlock need:
// 1. put block db
// 2. update chain
// 3. update balance
func WriteBlock(block types.Block, commonHash crypto.CommonHash)  {
	currentChain := GetChainCopy()
	log.Println("block number is ",block.Number)
	block.ParentHash = currentChain.LatestBlockHash
	block.BlockHash = block.Hash(commonHash).Bytes()
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, block.BlockHash, block)
	if err != nil {
		log.Fatal(err)
	}
	UpdateChain(block.BlockHash,false)
	balance, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("final number",currentChain.Height)
	fmt.Println("final hash",currentChain.LatestBlockHash)
	balance.UpdateDBBalance(&block)
}