package db_utils

import (
	"os"
	"strconv"
	"time"

	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"

	"github.com/golang/protobuf/proto"
)

// ==========================================================
// Public functions that invoked by outer service
// ==========================================================

// PersistBlock persists a block, using param to control whether flush to disk immediately.
func PersistBlock(batch db.Batch, block *types.Block, flush bool, sync bool, extra ...interface{}) (error, []byte) {
	if hyperdb.IfLogStatus() {
		go blockTime(block)
	}
	// Check pointer value
	if block == nil || batch == nil {
		return ErrEmptyPointer, nil
	}

	// Encapsulates block for specify block structure version
	err, data := encapsulateBlock(block, extra...)
	if err != nil {
		return err, nil
	}

	// Batch-Put the block in db
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

// GetBlock retrieves block with block hash.
func GetBlock(namespace string, key []byte) (*types.Block, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetBlockFunc(db, key)
}

// GetBlockFunc retrieves block with specific db.
func GetBlockFunc(db db.Database, key []byte) (*types.Block, error) {
	var (
		wrapper types.BlockWrapper
		block   types.Block
	)

	// Get the BlockWrapper by given key
	key = append(BlockPrefix, key...)
	data, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	// Unmarshal the BlockWrapper
	err = proto.Unmarshal(data, &wrapper)
	if err != nil {
		return &block, err
	}

	// Unmarshal the block
	data = wrapper.Block
	err = proto.Unmarshal(data, &block)

	return &block, err
}

// GetBlockByNumber retrieves block via block number.
func GetBlockByNumber(namespace string, blockNumber uint64) (*types.Block, error) {
	hash, err := getBlockHash(namespace, blockNumber)
	if err != nil {
		return nil, err
	}
	return GetBlock(namespace, hash)
}

// GetLatestBlock retrieves current head block.
func GetLatestBlock(namespace string) (*types.Block, error) {
	height := GetHeightOfChain(namespace)
	return GetBlockByNumber(namespace, height)
}

// DeleteBlockByNum deletes block data and block.num <---> block.hash
func DeleteBlockByNum(namepspace string, batch db.Batch, blockNum uint64, flush, sync bool) error {
	hash, err := getBlockHash(namepspace, blockNum)
	if err != nil {
		return err
	}
	return deleteBlock(namepspace, batch, hash, flush, sync)
}

// IsGenesisFinish checks whether genesis block has been mined into blockchain
func IsGenesisFinish(namespace string) bool {
	logger := common.GetLogger(namespace, "db_utils")
	tag, err := GetGenesisTag(namespace)
	if err != nil {
		return false
	}
	logger.Notice("tag: ", tag)
	_, err = GetBlockByNumber(namespace, tag)
	if err != nil {
		logger.Warning("missing genesis block")
		return false
	} else {
		return true
	}
}

// ==========================================================
// Private functions that invoked by inner service
// ==========================================================

// encapsulateBlock encapsulates block with a wrapper for specify block structure version.
func encapsulateBlock(block *types.Block, extra ...interface{}) (error, []byte) {
	var (
		blkVersion string = BlockVersion
		txVersion  string = TransactionVersion
	)

	if block == nil {
		return ErrEmptyPointer, nil
	}

	// Parse block and transaction version
	if len(extra) >= 1 {
		// parse version
		if tmp, ok := extra[0].(string); ok {
			blkVersion = tmp
		}
	}
	if len(extra) >= 2 {
		// parse version
		if tmp, ok := extra[1].(string); ok {
			txVersion = tmp
		}
	}
	block.Version = []byte(blkVersion)

	// Change every transaction in block to a specific version
	for _, tx := range block.Transactions {
		tx.Version = []byte(txVersion)
	}
	data, err := proto.Marshal(block)
	if err != nil {
		return err, nil
	}

	wrapper := &types.BlockWrapper{
		BlockVersion: []byte(blkVersion),
		Block:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		return err, nil
	}

	return nil, data
}

// getBlockHash retrieves block hash with related block number.
func getBlockHash(namespace string, blockNumber uint64) ([]byte, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return getBlockHashFunc(db, blockNumber)
}

// getBlockHashFunc retrieves block with specific db.
func getBlockHashFunc(db db.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(BlockNumPrefix, keyNum...))
}

// getBlockByNumberFunc retrieves block via block number with specific db.
func getBlockByNumberFunc(db db.Database, blockNumber uint64) (*types.Block, error) {
	hash, err := getBlockHashFunc(db, blockNumber)
	if err != nil {
		return nil, err
	}
	return GetBlockFunc(db, hash)
}

// deleteBlock deletes a block via hash.
func deleteBlock(namespace string, batch db.Batch, key []byte, flush, sync bool) error {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return err
	}
	return deleteBlockFunc(db, batch, key, flush, sync)
}

// deleteBlockFunc deletes a block via block hash with specific db.
func deleteBlockFunc(db db.Database, batch db.Batch, key []byte, flush, sync bool) error {
	// remove number <-> hash associate
	blk, err := GetBlockFunc(db, key)
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

// deleteBlockByNumFunc deletes a block via block num with specific db.
func deleteBlockByNumFunc(db db.Database, batch db.Batch, blockNum uint64, flush, sync bool) error {
	hash, err := getBlockHashFunc(db, blockNum)
	if err != nil {
		return err
	}
	return deleteBlockFunc(db, batch, hash, flush, sync)
}

// blockTime is for metric
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
