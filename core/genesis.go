// fetcher implements block operate
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-08-29
package core

import (
	"hyperchain/core/types"
	"io/ioutil"
	"encoding/json"

	"hyperchain/common"

	"hyperchain/hyperdb"

	"hyperchain/crypto"
	"time"
	"encoding/hex"
	//"hyperchain/core/vm/api"

)


func CreateInitBlock(filename string)  {
	log.Info("genesis start")

	if(GetHeightOfChain()>0){
		log.Info("already genesis")
		return
	}
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
		log.Error("ReadFile: ", err.Error())
		return
	}

	if err := json.Unmarshal(bytes, &genesis); err != nil {
		log.Error("Unmarshal: ", err.Error())
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

		balanceIns.PutCacheBalance(common.HexToAddress(addr),[]byte(account))
		balanceIns.PutDBBalance(common.HexToAddress(addr),[]byte(account))


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



	log.Debug("构造创世区块")
	err = PutBlock(db, block.BlockHash, &block)
	// write transaction
	//PutTransactions(db, commonHash, block.Transactions)
	if err != nil {
		log.Fatal(err)
	}
	UpdateChain(&block,true)
	log.Info("current chain block number is",GetChainCopy().Height)

}

// WriteBlock need:
// 1. Put block into db
// 2. Put transactions in block into db  (-- cancel --)
// 3. Update chain
// 4. Update balance
func WriteBlock(block *types.Block, commonHash crypto.CommonHash,commitTime int64)  {

	log.Info("block number is ",block.Number)
	currentChain := GetChainCopy()
	block.ParentHash = currentChain.LatestBlockHash
	block.BlockHash = block.Hash(commonHash).Bytes()
	block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, block.BlockHash, block)
	// write transaction
	//PutTransactions(db, commonHash, block.Transactions)
	if err != nil {
		log.Fatal(err)
	}
	UpdateChain(block, false)
	balance, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}

	newChain := GetChainCopy()
	log.Notice("Block number",newChain.Height)
	log.Notice("Block hash",hex.EncodeToString(newChain.LatestBlockHash))
	balance.UpdateDBBalance(block)
	// update our stateObject and statedb to blockchain
	ExecBlock(block)
	CommitStatedbToBlockchain()
}