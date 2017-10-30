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
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
)

var (
	LdbBlockCacheCapacity            = "database.leveldb.BlockCacheCapacity"
	LdbBlockRestartInterval          = 16
	LdbBlockSize                     = "database.leveldb.BlockSize"
	LdbCompactionExpandLimitFactor   = 25
	LdbCompactionGPOverlapsFactor    = 10
	LdbCompactionL0Trigger           = 4
	LdbCompactionSourceLimitFactor   = 1
	LdbCompactionTableSize           = 2 * MiB
	LdbCompactionTableSizeMultiplier = 1.0
	LdbCompactionTotalSize           = 10 * MiB
	LdbCompactionTotalSizeMultiplier = 10.0
	LdbCompressionType               = LdbSnappyCompression
	LdbIteratorSamplingRate          = 1 * MiB
	LdbMaxMemCompationLevel          = 2
	LdbNumLevel                      = 7
	LdbOpenFilesCacheCapacity        = 500
	LdbWriteBuffer                   = "database.leveldb.WriteBuffer"
	LdbWriteL0PauseTrigger           = "database.leveldb.WriteL0PauseTrigger"
	LdbWriteL0SlowdownTrigger        = "database.leveldb.WriteL0SlowdownTrigger"
)

type LdbCompression uint

const (
	LdbDefaultCompression LdbCompression = iota
	LdbNoCompression
	LdbSnappyCompression
	LdbnCompression
)

func GetLdbDefaultConfig() *opt.Options {
	return &opt.Options{}
}

func GetLdbConfig() *opt.Options {
	var config *common.Config

	opt := GetLdbDefaultConfig()

	opt.BlockSize = int(config.GetBytes(LdbBlockSize))
	opt.BlockCacheCapacity = int(config.GetBytes(LdbBlockCacheCapacity))
	opt.WriteBuffer = int(config.GetBytes(LdbWriteBuffer))
	opt.WriteL0PauseTrigger = config.GetInt(LdbWriteL0PauseTrigger)
	opt.WriteL0SlowdownTrigger = config.GetInt(LdbWriteL0SlowdownTrigger)

	return opt
}
