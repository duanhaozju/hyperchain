//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"hyperchain/hyperdb"
	"time"

	"io"
	"os"
	"path/filepath"
	"strconv"
	"errors"
	"fmt"
)

// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(blockNumber uint64, millTime int64) (int64, float64) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := GetBlockHash(db, blockNumber)
	block, err := GetBlock(db, blockHash)
	if err != nil {
		log.Error("Block not existed in database, err msg:", err.Error())
		return 0,0
	}
	var count int64 = 0
	for _, trans := range block.Transactions {
		if block.WriteTime-trans.Timestamp <= millTime*int64(time.Millisecond) {
			count++
		}
	}
	percent := float64(count) / 100

	//fmt.Println("block number is",block.Number)
	//if block.Transactions!=nil{
	//	fmt.Println("batch time is ",(block.Timestamp-block.Transactions[0].TimeStamp)/int64(time.Millisecond))
	//}

	/*fmt.Println("commit time is ",(block.CommitTime-block.Timestamp)/int64(time.Millisecond))
	fmt.Println("write time is ",(block.WriteTime-block.CommitTime)/ int64(time.Millisecond))
	fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))*/
	//fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))
	return count, percent
}

//CalcCommitAVGTime calculates block average commit time
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcCommitBatchAVGTime(from, to uint64) (int64, int64) {
	if from > to {
		log.Error("from less than to")
		return -1, -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var commit int64 = 0
	var batch int64 = 0
	for i := from; i <= to; i++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1, -1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1, -1
		}
		commit += block.CommitTime - block.Timestamp
		if block.Transactions != nil {
			batch += block.Timestamp - block.Transactions[0].Timestamp
		}

	}
	num := int64(to - from + 1)
	return commit / (num) / int64(time.Millisecond), batch / (num) / int64(time.Millisecond)

}
func CalTransactionSum() uint64 {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum uint64 = 0
	height := GetHeightOfChain()
	for i := uint64(1); i <= height; i++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return uint64(0)
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return uint64(0)
		}
		tmp := uint64(len(block.Transactions))
		//log.Info("block tx number is:",tmp)
		sum += tmp
	}
	return sum
}

// CalcResponseAVGTime calculate response avg time of blocks
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcResponseAVGTime(from, to uint64) int64 {
	if from > to && to != 0 {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	var length int64 = 0
	for i := from; i <= to; i++ {
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
		sum += (block.WriteTime - block.Timestamp) * int64(len(block.Transactions))
		length += int64(len(block.Transactions))
	}

	if length == 0 {
		return 0
	} else {
		return sum / (length * int64(time.Millisecond))
	}
}

func CalcBlockAVGTime(from, to uint64) int64 {
	if from > to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	for i := from; i <= to; i++ {
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
		sum += (block.WriteTime - block.Timestamp) / int64(time.Millisecond)
	}
	num := int64(to - from + 1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
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
	for i := from; i <= to; i++ {
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
		sum += (block.EvmTime - block.WriteTime) / int64(time.Microsecond)

	}
	num := int64(to - from + 1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
	}

}

func CalBlockGenerateAvgTime(from, to uint64) (int64,error){
	if from > to && to != 0 {
		log.Error("from less than to")
		return -1,errors.New("from less than to")
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error(err)
		return -1,err
	}
	var sum int64 = 0
	var length int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(db, i)
		if err != nil {
			log.Error(err)
			return -1,err
		}
		sum += (block.WriteTime - block.Timestamp) / int64(time.Millisecond)
		length++
	}

	return sum / length, nil
}

func CalBlockGPS(begin, end int64) (error, string) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	height := GetHeightOfChain()

	var s string
	s = s + "start time:" + time.Unix(0, begin).Format("2006-01-02 15:04:05") + ";"
	s = s + "end time:" + time.Unix(0, end).Format("2006-01-02 15:04:05") + ";"

	// calculate tps
	var blockCounter float64 = 0
	var txCounter float64 = 0
	var blockNum  uint64 = 0
	for i := uint64(1); i <= height; i++ {
		block, err := GetBlockByNumber(db, i)
		if err != nil {
			log.Error("Block not existed", err.Error())
			return err, ""
		}
		if block.WriteTime > end {
			break
		}
		if block.WriteTime > begin {
			txCounter += float64(len(block.Transactions))
			blockCounter += 1
			blockNum += 1
		}
	}
	s = s + "total block:" + strconv.FormatUint(blockNum, 16) + ";"
	blockCounter = blockCounter / (float64(end - begin) * 1.0 / float64(time.Second.Nanoseconds()))
	txCounter = txCounter / (float64(end - begin) * 1.0 / float64(time.Second.Nanoseconds()))
	s = s + "blocks per second:" + strconv.FormatFloat(blockCounter, 'f', 2, 32) + ";"
	s = s + "tps:" + strconv.FormatFloat(txCounter, 'f', 2, 32)
	return nil, s
	//for i := uint64(1); i <= height; i++ {
	//	block, err := GetBlockByNumber(db, i)
	//	if err != nil {
	//		log.Error("Block not existed", err.Error())
	//		return err
	//	}
	//	//println(time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),"********",block.Number)
	//	endSec := time.Unix(block.WriteTime/int64(time.Second), 0).Second()
	//	if block.WriteTime >= startTime && endSec-startSec == 0 {
	//		count++
	//		if i == height {
	//			current := time.Unix(0, startTime).Format("2006-01-02 15:04:05")
	//			s = current + ":" + strconv.Itoa(count) + " blocks generated" + "\n"
	//			content = append(content, s)
	//		}
	//		continue
	//	}
	//	current := time.Unix(0, startTime).Format("2006-01-02 15:04:05")
	//	s = current + ":" + strconv.Itoa(count) + " blocks generated" + "\n"
	//	content = append(content, s)
	//	flag = false
	//	if flag == false {
	//		startTime = block.WriteTime
	//		startSec = time.Unix(startTime/int64(time.Second), 0).Second()
	//		count = 1
	//		flag = true
	//		if i == height {
	//			s = current + ":" + strconv.Itoa(count) + " blocks generated" + "\n"
	//			content = append(content, s)
	//		}
	//	}
	//
	//}
	//path := "/tmp/hyperchain/cache/statis/block_time_statis"
	//return storeData(path, content)
	//return nil,
}

func GetBlockWriteTime(begin, end int64) (error, []string) {
	var ctx []string
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err, nil
	}
	for i := uint64(begin); i <= uint64(end); i++ {
		block, err := GetBlockByNumber(db, i)
		if err != nil {
			log.Error("Block not existed", err.Error())
			return err, nil
		}
		info := fmt.Sprintf("Block Number: %d, Write Time: %s", i, time.Unix(0, block.WriteTime).Format("2006-01-02 15:04:05"))
		ctx = append(ctx, info)
	}
	return nil, ctx

}
func storeData(file string, content []string) error {
	dir := filepath.Dir(file)
	_, err := os.Stat(dir)
	if !(err == nil || os.IsExist(err)) { //file exists
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	for _, d := range content {
		n, err := f.Write([]byte(d))
		if err == nil && n < len([]byte(d)) {
			err = io.ErrShortWrite
		}

	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
