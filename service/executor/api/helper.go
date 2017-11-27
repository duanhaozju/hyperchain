package api

import (
	"github.com/hyperchain/hyperchain/common"
	edb "github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/vm"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/hyperdb/db"

	"path"
	"fmt"
)

const (
	BLOCK               = "block"
	TRANSACTION         = "transaction"
	TRANSACTIONS        = "transactions"
	RECEIPT             = "receipt"
	DISCARDTXS          = "discard transactions"
	snapshotManifestPath  = "executor.archive.snapshot_manifest"

	CONTRACT                = "contract"
)

var (
	db_not_found_error = db.DB_NOT_FOUND.Error()
)

type BatchArgs struct {
	Hashes  []common.Hash `json:"hashes"`
	Numbers []BlockNumber `json:"numbers"`
	IsPlain bool		  `json:"isPlain"`			// whether returns block excluding transactions or not
}

type IntervalArgs struct {
	From         *BlockNumber    `json:"from"` // start block number
	To           *BlockNumber    `json:"to"`   // end block number
	ContractAddr *common.Address `json:"address"`
	MethodID     string          `json:"methodID"`
	IsPlain	 bool			 	 `json:"isPlain"`	// whether returns block excluding transactions or not
}

func substr(str string, start int, end int) string {
	rs := []rune(str)

	return string(rs[start:end])
}

// NewStateDb creates a new state db from latest block root.
func NewStateDb(conf *common.Config, namespace string) (vm.Database, error) {
	chain, err := edb.GetChain(namespace)
	if err != nil {
		return nil, err
	}

	height := chain.Height
	latestBlk, err := edb.GetBlockByNumber(namespace, height)
	if err != nil {
		return nil, err
	}
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	archiveDb, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_ARCHIVE)
	if err != nil {
		return nil, err
	}
	return state.New(common.BytesToHash(latestBlk.MerkleRoot), db, archiveDb, conf, height)
}

// NewSnapshotStateDb creates a new state db from given root.
func NewSnapshotStateDb(conf *common.Config, filterId string, merkleRoot []byte, height uint64, namespace string) (vm.Database, func(), error) {
	db, err := hyperdb.NewDatabase(conf, path.Join("snapshots", "SNAPSHOT_"+filterId), hyperdb.GetDatabaseType(conf), namespace)
	if err != nil {
		return nil, nil, err
	}
	closer := func() {
		db.Close()
	}
	state, err := state.New(common.BytesToHash(merkleRoot), db, nil, conf, height)
	return state, closer, err
}

type IntervalTime struct {
	StartTime int64 `json:"startTime"`
	Endtime   int64 `json:"endTime"`
}

type intArgs struct {
	from uint64
	to   uint64
}

// prepareIntervalArgs checks whether arguments are valid.
// If the client sends BlockNumber "", it will be converted to 0.
// If client sends BlockNumber 0, error will be returned.
func prepareIntervalArgs(args IntervalArgs, namespace string) (*intArgs, error) {
	if args.From == nil || args.To == nil {
		return nil, &common.InvalidParamsError{Message: "missing params `from` or `to`"}
	} else if chain, err := edb.GetChain(namespace); err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	} else {
		latest := chain.Height
		from, err := args.From.BlockNumberToUint64(latest)
		if err != nil {
			return nil, &common.InvalidParamsError{Message: err.Error()}
		}
		to, err := args.To.BlockNumberToUint64(latest)
		if err != nil {
			return nil, &common.InvalidParamsError{Message: err.Error()}
		}

		if from > to || from < 1 || to < 1 {
			return nil, &common.InvalidParamsError{Message: fmt.Sprintf("invalid params, from = %v, to = %v", from, to)}
		} else {
			return &intArgs{from: from, to: to}, nil
		}
	}
}

// prepareBlockNumber converts type BlockNumber to uint64.
func prepareBlockNumber(n BlockNumber, namespace string) (uint64, error) {
	chain, err := edb.GetChain(namespace)
	if err != nil {
		return 0, &common.CallbackError{Message: err.Error()}
	}
	latest := chain.Height
	number, err := n.BlockNumberToUint64(latest)
	if err != nil {
		return 0, &common.InvalidParamsError{Message: err.Error()}
	}
	return number, nil
}

// preparePagingArgs checks whether paging arguments are valid.
func preparePagingArgs(args PagingArgs) (PagingArgs, error) {
	if args.PageSize == 0 {
		return PagingArgs{}, &common.InvalidParamsError{Message: "invalid params, '`ageSize` value shouldn't be 0"}
	} else if args.Separated%args.PageSize != 0 {
		return PagingArgs{}, &common.InvalidParamsError{Message: "invalid params, `separated` value should be a multiple of `pageSize` value"}
	} else if args.MaxBlkNumber == BlockNumber(0) || args.MinBlkNumber == BlockNumber(0) {
		return PagingArgs{}, &common.InvalidParamsError{Message: "invalid params, `minBlkNumber` or `maxBlkNumber` value shouldn't be 0"}
	} else if args.ContractAddr == nil {
		return PagingArgs{}, &common.InvalidParamsError{Message: "invalid params, missing params `address`"}
	}

	return args, nil
}

func GetManifestPath(conf *common.Config) string {
	return conf.GetString(snapshotManifestPath)
}
