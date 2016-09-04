package core

import (
	"hyperchain/hyperdb"
	"time"
	"fmt"
)

// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(blockNumber uint64, millTime int64) int64 {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := GetBlockHash(db, blockNumber)
	fmt.Println(blockHash)
	block, err := GetBlock(db, blockHash)
	var count int64 = 0
	for i, trans := range block.Transactions {

		fmt.Println(time.Unix(block.WriteTime/1e9, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Println(time.Unix(trans.TimeStamp/1e9, 0).Format("2006-01-02 03:04:05 PM"))

		fmt.Println(i , " : ", block.Timestamp - trans.TimeStamp)
		fmt.Println(i , " : ", block.WriteTime - trans.TimeStamp)
		if block.WriteTime - trans.TimeStamp <= millTime * int64(time.Millisecond) {
			count ++
		}
	}
	return count
}