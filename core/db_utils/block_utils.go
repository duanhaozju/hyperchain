package db_utils

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"os"
	"strconv"
	"time"
)

// PersistBlock - persist a block, using param to control whether flush to disk immediately.
func PersistBlock(batch db.Batch, block *types.Block, flush bool, sync bool, extra ...interface{}) (error, []byte) {
	if hyperdb.IfLogStatus() {
		go blockTime(block)
	}
	// check pointer value
	if block == nil || batch == nil {
		return EmptyPointerErr, nil
	}
	err, data := encapsulateBlock(block, extra...)
	if err != nil {
		return err, nil
	}
	if err := batch.Put(append(BlockPrefix, block.BlockHash...), data); err != nil {
		return err, nil
	}

	// save number <-> hash associate
	keyNum := strconv.FormatUint(block.Number, 10)
	if err := batch.Put(append(BlockNumPrefix, []byte(keyNum)...), block.BlockHash); err != nil {
		return err, nil
	}
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

// encapsulateBlock - encapsulate block with a wrapper for specify block structure version.
func encapsulateBlock(block *types.Block, extra ...interface{}) (error, []byte) {
	version := BlockVersion
	if block == nil {
		return EmptyPointerErr, nil
	}
	if len(extra) > 0 {
		// parse version
		if tmp, ok := extra[0].(string); ok {
			version = tmp
		}
	}
	block.Version = []byte(version)
	data, err := proto.Marshal(block)
	if err != nil {
		return err, nil
	}
	wrapper := &types.BlockWrapper{
		BlockVersion: []byte(version),
		Block:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		return err, nil
	}
	return nil, data
}

// GetBlockHash - retrieve block hash with related block number.
func GetBlockHash(namespace string, blockNumber uint64) ([]byte, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetBlockHashFunc(db, blockNumber)
}

// GetBlockHashFunc - retrieve block with specific db.
func GetBlockHashFunc(db db.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(BlockNumPrefix, keyNum...))
}

// GetBlock - retrieve block with block hash.
func GetBlock(namespace string, key []byte) (*types.Block, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetBlockFunc(db, key)
}

// GetBlockFunc - retrieve block with specific db.
func GetBlockFunc(db db.Database, key []byte) (*types.Block, error) {
	var wrapper types.BlockWrapper
	var block types.Block
	key = append(BlockPrefix, key...)
	data, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(data, &wrapper)
	if err != nil {
		return &block, err
	}
	data = wrapper.Block
	err = proto.Unmarshal(data, &block)
	return &block, err
}

// GetBlockByNumber - retrieve block via block number.
func GetBlockByNumber(namespace string, blockNumber uint64) (*types.Block, error) {
	hash, err := GetBlockHash(namespace, blockNumber)
	if err != nil {
		return nil, err
	}
	return GetBlock(namespace, hash)
}

// GetLatestBlock get current head block.
func GetLatestBlock(namespace string) (*types.Block, error) {
	height := GetHeightOfChain(namespace)
	return GetBlockByNumber(namespace, height)
}

// GetBlockByNumberFunc - retrieve block via block number with specific db.
func GetBlockByNumberFunc(db db.Database, blockNumber uint64) (*types.Block, error) {
	hash, err := GetBlockHashFunc(db, blockNumber)
	if err != nil {
		return nil, err
	}
	return GetBlockFunc(db, hash)
}

// DeleteBlock - delete a block via hash.
func DeleteBlock(namespace string, batch db.Batch, key []byte, flush, sync bool) error {
	// remove number <-> hash associate
	blk, err := GetBlock(namespace, key)
	if err != nil {
		return err
	}
	keyNum := strconv.FormatUint(blk.Number, 10)
	err = batch.Delete(append(BlockNumPrefix, []byte(keyNum)...))
	if err != nil {
		return err
	}
	// remove block
	keyFact := append(BlockPrefix, key...)
	err = batch.Delete(keyFact)
	if err != nil {
		return err
	}
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// DeleteBlockByNum - delete block data and block.num<--->block.hash
func DeleteBlockByNum(namepspace string, batch db.Batch, blockNum uint64, flush, sync bool) error {
	hash, err := GetBlockHash(namepspace, blockNum)
	if err != nil {
		return err
	}
	return DeleteBlock(namepspace, batch, hash, flush, sync)
}

// IsGenesisFinish - check whether genesis block has been mined into blockchain
func IsGenesisFinish(namespace string) bool {
	_, err := GetBlockByNumber(namespace, 0)
	if err != nil {
		logger := common.GetLogger(namespace, "db_utils")
		logger.Warning("missing genesis block")
		return false
	} else {
		return true
	}
}

// BlockTime - for metric
func blockTime(block *types.Block) {
	times := block.WriteTime - block.Timestamp
	f, err := os.OpenFile(hyperdb.GetLogPath(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		//logger.Notice("db.logger file create failed. err: " + err.Error())
	} else {
		n, _ := f.Seek(0, os.SEEK_END)
		currentTime := time.Now().Local()
		newFormat := currentTime.Format("2006-01-02 15:04:05.000")
		str := strconv.FormatUint(block.Number, 10) + "#" + newFormat + "#" + strconv.FormatInt(times, 10) + "\n"
		_, err = f.WriteAt([]byte(str), n)
		f.Close()
	}
}
