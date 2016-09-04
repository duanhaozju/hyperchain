package core

import (
	"hyperchain/hyperdb"
	"time"
)

// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(blockNumber uint64, millTime int64) int64 {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := GetBlockHash(db, blockNumber)
	block, err := GetBlock(db, blockHash)
	var count int64 = 0
	for _, trans := range block.Transactions {
		if block.WriteTime - trans.TimeStamp <= millTime * int64(time.Millisecond) {
			count ++
		}
	}
	return count
}

// CalcResponseAVGTime calculate response avg time of blocks
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcResponseAVGTime(from, to uint64) int64 {
	if from < to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1
		}
		for _, trans := range block.Transactions {
			sum += block.WriteTime - trans.TimeStamp
		}
	}

	return sum / (int64(to - from + 1) * 500 * int64(time.Millisecond))
}