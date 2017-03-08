package db_utils

import (
	"time"
	"fmt"
	"path/filepath"
	"os"
	"io"
	"errors"
	"strconv"
)
// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(namespace string, blockNumber uint64, millTime int64) (int64, float64) {
	block, err := GetBlockByNumber(namespace, blockNumber)
	if err != nil {
		logger.Errorf("[Namespace = %s] block not existed in database, err msg : ", namespace, err.Error())
		return 0, 0
	}
	var count int64 = 0
	for _, trans := range block.Transactions {
		if block.WriteTime-trans.Timestamp <= millTime*int64(time.Millisecond) {
			count++
		}
	}
	percent := float64(count) / 100
	return count, percent
}

//CalcCommitAVGTime calculates block average commit time
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcCommitBatchAVGTime(namespace string, from, to uint64) (int64, int64) {
	if from > to {
		logger.Error("from less than to")
		return -1, -1
	}
	var commit int64 = 0
	var batch int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
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
func CalTransactionSum(namespace string) uint64 {
	var sum uint64 = 0
	height := GetHeightOfChain(namespace)
	for i := uint64(1); i <= height; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
			return uint64(0)
		}
		tmp := uint64(len(block.Transactions))
		sum += tmp
	}
	return sum
}

// CalcResponseAVGTime calculate response avg time of blocks
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcResponseAVGTime(namespace string, from, to uint64) int64 {
	if from > to && to != 0 {
		logger.Error("from less than to")
		return -1
	}
	var sum int64 = 0
	var length int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
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

func CalcBlockAVGTime(namespace string, from, to uint64) int64 {
	if from > to {
		logger.Error("from less than to")
		return -1
	}
	var sum int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
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

func CalcEvmAVGTime(namespace string, from, to uint64) int64 {
	if from > to {
		logger.Error("from less than to")
		return -1
	}
	var sum int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
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

func CalBlockGenerateAvgTime(namespace string, from, to uint64) (int64, error) {
	if from > to && to != 0 {
		logger.Error("from less than to")
		return -1, errors.New("from less than to")
	}
	var sum int64 = 0
	var length int64 = 0
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error(err)
			return -1, err
		}
		sum += (block.WriteTime - block.Timestamp) / int64(time.Millisecond)
		length++
	}

	return sum / length, nil
}

func CalBlockGPS(namespace string, begin, end int64) (error, string) {
	height := GetHeightOfChain(namespace)

	var s string
	s = s + "start time:" + time.Unix(0, begin).Format("2006-01-02 15:04:05") + ";"
	s = s + "end time:" + time.Unix(0, end).Format("2006-01-02 15:04:05") + ";"

	// calculate tps
	var blockCounter float64 = 0
	var txCounter float64 = 0
	var blockNum uint64 = 0
	for i := uint64(1); i <= height; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error("Block not existed", err.Error())
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
	s = s + "total block:" + strconv.FormatUint(blockNum, 10) + ";"
	blockCounter = blockCounter / (float64(end-begin) * 1.0 / float64(time.Second.Nanoseconds()))
	txCounter = txCounter / (float64(end-begin) * 1.0 / float64(time.Second.Nanoseconds()))
	s = s + "blocks per second:" + strconv.FormatFloat(blockCounter, 'f', 2, 32) + ";"
	s = s + "tps:" + strconv.FormatFloat(txCounter, 'f', 2, 32)
	return nil, s
}

func GetBlockWriteTime(namespace string, begin, end int64) (error, []string) {
	var ctx []string
	for i := uint64(begin); i <= uint64(end); i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger.Error("Block not existed", err.Error())
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
