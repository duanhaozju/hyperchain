package chain

import (
	"errors"

	"hyperchain/common"
	"hyperchain/hyperdb"

	"github.com/op/go-logging"
    "hyperchain/common/service/server"
)

var (
	ErrEmptyPointer      = errors.New("nil pointer")
	ErrNotFindTxMeta     = errors.New("not find tx meta")
	ErrNotFindBlock      = errors.New("not find block")
	ErrOutOfSliceRange   = errors.New("out of slice(transactions in block) range")
	ErrFromSmallerThanTo = errors.New("from smaller than to")
)

const (
	BlockVersion       = "1.3"
	ReceiptVersion     = "1.3"
	TransactionVersion = "1.3"
)

var (
	ChainKey = []byte("chain-key-")

	// Data item prefixes
	ReceiptsPrefix  = []byte("receipts-")
	InvalidTxPrefix = []byte("invalidTx-")
	BlockPrefix     = []byte("block-")
	BlockNumPrefix  = []byte("blockNum-")
	TxMetaSuffix    = []byte{0x01}

	JournalPrefix  = []byte("-journal")
	SnapshotPrefix = []byte("-snapshot")
)

// InitDBForNamespace inits the database with given namespace
func InitDBForNamespace(conf *common.Config, namespace string, is *server.InternalServer) error {
	err := hyperdb.InitDatabase(conf, namespace)
	if err != nil {
		return err
	}
    if conf.GetBool(common.EXECUTOR_EMBEDDED) {
        InitializeChain(namespace)
    } else {
        InitializeRemoteChain(is, namespace)
    }
	return err
}

func InitExecutorDBForNamespace(conf *common.Config, namespace string) error {
    err := hyperdb.InitBlockDatabase(conf, namespace)
    if err != nil {
        return err
    }
    err = hyperdb.InitArchiveDatabase(conf, namespace)
    if err != nil {
        return err
    }
    InitializeChain(namespace)
    return err
}


// logger returns the logger with given namespace
func logger(namespace string) *logging.Logger {
	return common.GetLogger(namespace, "core/ledger")
}
