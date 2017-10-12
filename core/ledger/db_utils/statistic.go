package db_utils

import (
	"fmt"
	"strconv"
	"time"
)

// CalcCommitAVGTime calculates block avg commit and batch time of blocks,
// whose blockNumber from 'from' to 'to' ( ['from', 'to'] ).
// Returns : avg Millisecond.
func CalcCommitBatchAVGTime(namespace string, from, to uint64) (int64, int64) {
	var commit, batch, num int64

	// Return if from > to
	if from > to {
		logger(namespace).Error(ErrFromSmallerThanTo.Error())
		return -1, -1
	}

	commit, batch = 0, 0
	// Calculate blocks in ['from', 'to']
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(err)
			return -1, -1
		}
		commit += block.CommitTime - block.Timestamp
		// Choose the first transaction in the batch as the base point
		// TODO: however, the transactions in batch are not ordered by time
		if block.Transactions != nil {
			batch += block.Timestamp - block.Transactions[0].Timestamp
		}
	}

	num = int64(to - from + 1)
	return commit / num / int64(time.Millisecond), batch / num / int64(time.Millisecond)
}

// CalcResponseAVGTime calculates avg response time of blocks,
// whose blockNumber from 'from' to 'to' ( ['from', 'to'] ).
// Return : avg Millisecond.
func CalcResponseAVGTime(namespace string, from, to uint64) int64 {
	var sum, length int64

	// Return if from > to
	if from > to && to != 0 {
		logger(namespace).Error(ErrFromSmallerThanTo.Error())
		return -1
	}

	sum, length = 0, 0
	// Calculate blocks in ['from', 'to']
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(err)
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

// CalcEvmAVGTime calculates avg evm's running time of blocks,
// whose blockNumber from 'from' to 'to' ( ['from', 'to'] ).
// Return : avg Millisecond.
func CalcEvmAVGTime(namespace string, from, to uint64) int64 {
	var (
		sum int64 = 0
		num int64
	)

	// Return if from > to
	if from > to {
		logger(namespace).Error(ErrFromSmallerThanTo.Error())
		return -1
	}

	// Calculate blocks in ['from', 'to']
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(err)
			return -1
		}
		sum += (block.EvmTime - block.WriteTime) / int64(time.Microsecond)
	}

	num = int64(to - from + 1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
	}

}

// CalBlockGenerateAvgTime calculates avg generate time of blocks,
// whose blockNumber from 'from' to 'to' ( ['from', 'to'] ).
// Return : avg Millisecond, error.
func CalBlockGenerateAvgTime(namespace string, from, to uint64) (int64, error) {
	var sum, length int64

	// Return if from > to
	if from > to && to != 0 {
		logger(namespace).Error(ErrFromSmallerThanTo.Error())
		return -1, ErrFromSmallerThanTo
	}

	sum, length = 0, 0
	// Calculate blocks in ['from', 'to']
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(err)
			return -1, err
		}
		sum += (block.WriteTime - block.Timestamp) / int64(time.Millisecond)
		length++
	}

	return sum / length, nil
}

// GetBlockWriteTime calculates avg write time of blocks,
// whose blockNumber from 'from' to 'to' ( ['from', 'to'] ).
// Return : description, error.
func GetBlockWriteTime(namespace string, from, to uint64) ([]string, error) {
	var ctx []string
	// Calculate blocks in ['from', 'to']
	for i := from; i <= to; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(ErrNotFindBlock.Error(), err.Error())
			return nil, ErrNotFindBlock
		}
		info := fmt.Sprintf("Block Number: %d, Write Time: %s", i, time.Unix(0, block.WriteTime).Format("2006-01-02 15:04:05"))
		ctx = append(ctx, info)
	}

	return ctx, nil
}

// CalBlockGPS calculates avg generate time of blocks,
// whose blockNumber from 'begin' to 'end' ( ['begin', 'end'] ).
// Return : description, error.
func CalBlockGPS(namespace string, begin, end int64) (string, error) {
	var (
		blockCounter float64 = 0
		txCounter    float64 = 0
		blockNum     uint64  = 0
		desc         string
	)

	height := GetHeightOfChain(namespace)
	tag, _ := GetGenesisTag(namespace)

	desc = desc + "start time:" + time.Unix(0, begin).Format("2006-01-02 15:04:05") + ";"
	desc = desc + "end time:" + time.Unix(0, end).Format("2006-01-02 15:04:05") + ";"

	// Calculate the tps [ GenesisTag, ChainHeight ]
	for i := tag; i <= height; i++ {
		block, err := GetBlockByNumber(namespace, i)
		if err != nil {
			logger(namespace).Error(ErrNotFindBlock.Error(), err.Error())
			return "", ErrNotFindBlock
		}
		if block.WriteTime > end {
			break
		}
		if block.WriteTime > begin {
			txCounter += float64(len(block.Transactions))
			blockCounter++
			blockNum++
		}
	}
	desc = desc + "total block:" + strconv.FormatUint(blockNum, 10) + ";"
	blockCounter = blockCounter / (float64(end-begin) * 1.0 / float64(time.Second.Nanoseconds()))
	txCounter = txCounter / (float64(end-begin) * 1.0 / float64(time.Second.Nanoseconds()))
	desc = desc + "blocks per second:" + strconv.FormatFloat(blockCounter, 'f', 2, 32) + ";"
	desc = desc + "tps:" + strconv.FormatFloat(txCounter, 'f', 2, 32)
	return desc, nil
}
