package core

import (
	"fmt"
	"hyperchain/core/types"
	"io/ioutil"
	"encoding/json"

	"hyperchain/common"

	"log"

)


func CreateInitBlock(filename string)  {

	type Genesis struct {

		Timestamp  int64
		ParentHash  string
		BlockHash  string
		Coinbase    string
		Number int64
		Alloc       map[string]int64




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

	for addr, account := range genesis["test1"].Alloc {
		//address := common.HexToAddress(addr)

		//value, err := strconv.ParseInt(account.Balance, 10, 64)
		/*fmt.Println(addr)
		fmt.Println([]byte(addr))
		fmt.Println(common.BytesToHash([]byte(addr)))
		fmt.Println(common.BytesToHash([]byte("0000000000000000000000000000000000000002")))
		*/balance:=types.Balance{
			AccountPublicKeyHash:[]byte(addr),
			Value:account,
		}
		if err!=nil{
			return
		}
		PutBalanceToMEM(balance)
		//.AddBalance(address, common.String2Big(account.Balance))


	}



	block := types.Block{
		ParentHash: common.FromHex(genesis["test1"].ParentHash),
		Timestamp:   genesis["test1"].Timestamp,
		BlockHash: common.FromHex(genesis["test1"].BlockHash),
		Number:   genesis["test1"].Number,
		//MerkleRoot:       "root",
	}



	log.Println("构造创世区块")

	UpdateChain(block.BlockHash)


	fmt.Print(GetBalanceFromMEM(common.BytesToAddress([]byte("0000000000000000000000000000000000000002"))).Value)



	//-- 初始初始化balance
	//core.UpdateBalance(block)
}
