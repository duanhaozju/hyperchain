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

	// database names
	DBNAME_ARCHIVE    = "database.dbname.archive"
	DBNAME_BLOCKCHAIN = "database.dbname.blockchain"
	DBNAME_CONSENSUS  = "database.dbname.consensus"

	// database storage directories
	DBPATH_ARCHIVE    = "database.dbpath.archive"
	DBPATH_BLOCKCHAIN = "database.dbpath.blockchain"
	DBPATH_CONSENSUS  = "database.dbpath.consensus"

	// database types
	LDB_DB         = 0001
	SUPER_LEVEL_DB = 0010
	MEMORY_DB      = 0011
)

const (
	LdbBlockCacheCapacity     = "database.leveldb.BlockCacheCapacity"
	LdbBlockSize              = "database.leveldb.BlockSize"
	LdbWriteBuffer            = "database.leveldb.WriteBuffer"
	LdbWriteL0PauseTrigger    = "database.leveldb.WriteL0PauseTrigger"
	LdbWriteL0SlowdownTrigger = "database.leveldb.WriteL0SlowdownTrigger"
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
