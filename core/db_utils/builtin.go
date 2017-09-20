package db_utils

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb"
)

var (
	EmptyPointerErr    = errors.New("nil pointer")
	MarshalErr         = errors.New("marshal failed")
	NotFindTxMetaErr   = errors.New("not find tx meta")
	NotFindBlockErr    = errors.New("not find block")
	OutOfSliceRangeErr = errors.New("out of slice(transactions in block) range")
)

const (
	BlockVersion       = "1.3"
	ReceiptVersion     = "1.3"
	TransactionVersion = "1.3"
	ChainVersion       = "1.3"
)

var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key-")
	BlockNumPrefix           = []byte("blockNum-")
	TxMetaSuffix             = []byte{0x01}

	JournalPrefix  = []byte("-journal")
	SnapshotPrefix = []byte("-snapshot")
	BloomPrefix    = []byte("bloom-")
)

func InitDBForNamespace(conf *common.Config, namespace string) error {
	err := hyperdb.InitDatabase(conf, namespace)
	if err != nil {
		return err
	}
	InitializeChain(namespace)
	return err
}

func logger(namespace string) *logging.Logger {
	return common.GetLogger(namespace, "core/db_utils")
}
