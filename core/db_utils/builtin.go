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
	BlockVersion       = "1.0"
	ReceiptVersion     = "1.0"
	TransactionVersion = "1.0"
	DefaultNamespace   = "Global"
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

// using to count the number of rollback transactions
var RollbackDataSum int = 0

func init() {
	logger = logging.MustGetLogger("db_utils")
}

func InitDB(conf *common.Config, namespace string, dbConfig string, port int) {
	hyperdb.SetDBConfig(dbConfig, strconv.Itoa(port))
	hyperdb.InitDatabase(conf, namespace)
	InitializeChain(DefaultNamespace)
}
