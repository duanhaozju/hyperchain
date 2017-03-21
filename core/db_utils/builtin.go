package db_utils

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"strconv"
)

var (
	logger          *logging.Logger // package-level logger
	EmptyPointerErr = errors.New("nil pointer")
	MarshalErr      = errors.New("marshal failed")
)

const (
	BlockVersion       = "1.2"
	ReceiptVersion     = "1.2"
	TransactionVersion = "1.2"
)

var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key-")
	BlockNumPrefix           = []byte("blockNum-")
	TxMetaSuffix             = []byte{0x01}
)

func init() {
	logger = logging.MustGetLogger("db_utils")
}

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
