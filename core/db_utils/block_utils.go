package db_utils

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"os"
	"strconv"
	"time"
)

// PersistBlock - persist a block, using param to control whether flush to disk immediately.
func PersistBlock(batch db.Batch, block *types.Block, flush bool, sync bool) (error, []byte) {
	if hyperdb.IfLogStatus() {
		go blockTime(block)
	}
	// check pointer value
	if block == nil || batch == nil {
		return EmptyPointerErr, nil
	}
	err, data := encapsulateBlock(block)
	if err != nil {
		logger.Errorf("wrapper block failed.")
		return err, nil
	}
	if err := batch.Put(append(BlockPrefix, block.BlockHash...), data); err != nil {
		logger.Error("Put block data into database failed! error msg, ", err.Error())
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
func encapsulateBlock(block *types.Block) (error, []byte) {
	if block == nil {
		return EmptyPointerErr, nil
	}
	block.Version = []byte(BlockVersion)
	data, err := proto.Marshal(block)
	if err != nil {
		logger.Errorf("marshal block failed, %s", err.Error())
		return err, nil
	}
	wrapper := &types.BlockWrapper{
		BlockVersion: []byte(BlockVersion),
		Block:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		logger.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}

// GetBlockHash - retrieve block hash with related block number.
func GetBlockHash(namespace string, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return db.Get(append(BlockNumPrefix, keyNum...))
}

// GetBlock - retrieve block with block hash.
func GetBlock(namespace string, key []byte) (*types.Block, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
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
		logger.Warning("missing genesis block")
		return false
	} else {
		return true
	}
}
func blockTime(block *types.Block) {
	time1 := block.WriteTime - block.Timestamp
	f, err1 := os.OpenFile(hyperdb.GetLogPath(), os.O_WRONLY|os.O_CREATE, 0644)
	if err1 != nil {
		logger.Notice("db.logger file create failed. err: " + err1.Error())
	} else {
		n, _ := f.Seek(0, os.SEEK_END)
		currentTime := time.Now().Local()
		newFormat := currentTime.Format("2006-01-02 15:04:05.000")
		str := strconv.FormatUint(block.Number, 10) + "#" + newFormat + "#" + strconv.FormatInt(time1, 10) + "\n"
		_, err1 = f.WriteAt([]byte(str), n)
		f.Close()
	}
}
