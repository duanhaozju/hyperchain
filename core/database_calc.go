package core

import (
	"hyperchain/hyperdb"
	"time"

	//"fmt"

)

// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(blockNumber uint64, millTime int64) (int64,float64){
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
	percent := float64(count)/500

	//fmt.Println("block number is",block.Number)
	//if block.Transactions!=nil{
	//	fmt.Println("batch time is ",(block.Timestamp-block.Transactions[0].TimeStamp)/int64(time.Millisecond))
	//}

	/*fmt.Println("commit time is ",(block.CommitTime-block.Timestamp)/int64(time.Millisecond))
	fmt.Println("write time is ",(block.WriteTime-block.CommitTime)/ int64(time.Millisecond))
	fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))*/
	//fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))
	return count,percent
}
//CalcCommitAVGTime calculates block average commit time
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcCommitBatchAVGTime(from,to uint64) (int64,int64) {
	if from > to {
		log.Error("from less than to")
		return -1,-1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var commit int64 = 0
	var batch int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1,-1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1,-1
		}
		commit += block.CommitTime - block.Timestamp
		if block.Transactions!=nil{
			batch += block.Timestamp - block.Transactions[0].TimeStamp
		}

	}
	num := int64(to-from+1)
	return commit/(num)/int64(time.Millisecond),batch/(num)/int64(time.Millisecond)

}

// CalcResponseAVGTime calculate response avg time of blocks
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcResponseAVGTime(from, to uint64) int64 {
	if from > to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	var length int64 = 0
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
		length += int64(len(block.Transactions))
	}

	if length == 0 {
		return 0
	} else {
		return sum / (length * int64(time.Millisecond))
	}


}


func CalcEvmAVGTime(from, to uint64) int64 {
	if from > to {
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
		sum+=(block.EvmTime-block.WriteTime)/ int64(time.Millisecond)
	}
	num := int64(to-from+1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
	}


}