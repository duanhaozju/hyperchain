//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/hyperchain/hyperchain/common"
)

const (
	LEVEL_DB_ROOT_DIR = "database.leveldb.root_dir"
	DB_TYPE           = "database.type"

	// database types
	LDB_DB    = 0001
	MEMORY_DB = 0010
)

const (
	// database indexes
	DBINDEX_BLOCKCHAIN int = iota
	DBINDEX_CONSENSUS
	DBINDEX_ARCHIVE
	DBINDEX_MAX

	// database names
	DBNAME_BLOCKCHAIN = "blockchain"
	DBNAME_CONSENSUS  = "consensus"
	DBNAME_ARCHIVE    = "archive"

	// database storage directories
	DBPATH_BLOCKCHAIN = "database.dbpath.blockchain"
	DBPATH_CONSENSUS  = "database.dbpath.consensus"
	DBPATH_ARCHIVE    = "database.dbpath.archive"
)

const (
	// whether encrypt data
	DATA_ENCRYPTION = "database.leveldb.data_encryption"

	// encryption function in secimpl requires the length of key to be 24
	ENCRYPTED_KEY_LEN = 24

	// to append the key if its length is less than "ENCRYPTED_KEY_LEN"
	// it doesn't matter what the value here is.
	ENCRYPTED_KEY_APPENDER = 55
)

const (
	LdbBlockCacheCapacity     = "database.leveldb.options.block_cache_capacity"
	LdbBlockSize              = "database.leveldb.options.block_size"
	LdbWriteBuffer            = "database.leveldb.options.write_buffer"
	LdbWriteL0PauseTrigger    = "database.leveldb.options.write_l0_pause_trigger"
	LdbWriteL0SlowdownTrigger = "database.leveldb.options.write_l0_slowdown_trigger"
)

func GetLdbDefaultConfig() *opt.Options {
	return &opt.Options{}
}

func GetLdbConfig(conf *common.Config) *opt.Options {
	opt := GetLdbDefaultConfig()

	opt.BlockSize = int(conf.GetBytes(LdbBlockSize))
	opt.BlockCacheCapacity = int(conf.GetBytes(LdbBlockCacheCapacity))
	opt.WriteBuffer = int(conf.GetBytes(LdbWriteBuffer))
	opt.WriteL0PauseTrigger = conf.GetInt(LdbWriteL0PauseTrigger)
	opt.WriteL0SlowdownTrigger = conf.GetInt(LdbWriteL0SlowdownTrigger)

	return opt
}
