package db_utils

import (
	"errors"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"strconv"
	"github.com/op/go-logging"
)

var (
	EmptyPointerErr = errors.New("nil pointer")
	MarshalErr      = errors.New("marshal failed")
)

const (
	BlockVersion       = "1.2"
	ReceiptVersion     = "1.2"
	TransactionVersion = "1.2"
	ChainVersion       = "1.2"
)

var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key-")
	BlockNumPrefix           = []byte("blockNum-")
	TxMetaSuffix             = []byte{0x01}

	JournalPrefix            = []byte("-journal")
	SnapshotPrefix           = []byte("-snapshot")
)

func InitDBForNamespace(conf *common.Config, namespace string) error{
	dbConfigPath := conf.GetString(common.DB_CONFIG_PATH)
	conf.MergeConfig(dbConfigPath)
	hyperdb.SetDBConfig(dbConfigPath, strconv.Itoa(conf.GetInt(common.C_NODE_ID)))
	err := hyperdb.InitDatabase(conf, namespace)
	if err != nil {
		return err
	}
	InitializeChain(namespace)
	return err
}

func logger(namespace string) *logging.Logger  {
	return common.GetLogger(namespace, "core/db_utils")
}
