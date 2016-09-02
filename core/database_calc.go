package core

import (
	"hyperchain/hyperdb"
	"log"
)

// CalcResponseCount calculate response count of a block for given blockNumber
func CalcResponseCount(blockNumber uint64, time int64) int64 {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := GetBlockHash(db, blockNumber)
	block, err := GetBlock(db, blockHash)
	var count int64 = 0
	for _, trans := range block.Transactions {
		if block.WriteTime - trans.TimeStamp <= time {
			count ++
		}
	}
	return count
}