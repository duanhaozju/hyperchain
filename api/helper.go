package api

import (
	"errors"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"hyperchain/crypto/hmEncryption"
	"hyperchain/hyperdb"
	"math/big"
	"path"
	"time"
)

const (
	rateLimitEnable = "flow.control.ratelimit.enable"

	transactionPeak     = "flow.control.ratelimit.txRatePeak"
	transactionFillRate = "flow.control.ratelimit.txFillRate"

	contractPeak     = "flow.control.ratelimit.contractRatePeak"
	contractFillRate = "flow.control.ratelimit.contractFillRate"

	paillpublickeyN       = "global.configs.hmpublickey.N"
	paillpublickeynsquare = "global.configs.hmpublickey.Nsquare"
	paillpublickeyG       = "global.configs.hmpublickey.G"
	snapshotManifestPath  = "executor.archive.snapshot_manifest"

	BLOCK				= "block"
	TRANSACTION         = "transaction"
	CONTRACT            = "contract"
	RECEIPT				= "receipt"

	DEFAULT_GAS      int64 = 100000000
	DEFAULT_GAS_PRICE int64 = 10000

	DISCARDTXS			= "discard transactions"
	DISCARDTX			= "discard transaction"

)

// getRateLimitEnable returns rate limit switch value.
func getRateLimitEnable(conf *common.Config) bool {
	return conf.GetBool(rateLimitEnable)
}

func getManifestPath(conf *common.Config) string {
	return conf.GetString(snapshotManifestPath)
}

// getRateLimitPeak returns rate limit peak value.
func getRateLimitPeak(namespace string, conf *common.Config, choice string) int64 {
	log := common.GetLogger(namespace, "api")
	switch choice {
	case TRANSACTION:
		return conf.GetInt64(transactionPeak)
	case CONTRACT:
		return conf.GetInt64(contractPeak)
	default:
		log.Errorf("no choice specified. %s or %s", TRANSACTION, CONTRACT)
		return 0
	}
}

// getFillRate returns rate limit fill speed.
func getFillRate(namespace string, conf *common.Config, choice string) (time.Duration, error) {
	log := common.GetLogger(namespace, "api")
	switch choice {
	case TRANSACTION:
		return time.ParseDuration(conf.GetString(transactionFillRate))
	case CONTRACT:
		return time.ParseDuration(conf.GetString(contractFillRate))
	default:
		log.Errorf("no choice specified. %s or %s", TRANSACTION, CONTRACT)
		return time.Duration(0), errors.New("no choice specified in get fill rate")
	}
}

// getPaillierPublickey returns public key for hmEncryption.
func getPaillierPublickey(config *common.Config) hmEncryption.PaillierPublickey {
	bigN := new(big.Int)
	bigNsquare := new(big.Int)
	bigG := new(big.Int)
	n, _ := bigN.SetString(config.GetString(paillpublickeyN), 10)
	nsquare, _ := bigNsquare.SetString(config.GetString(paillpublickeynsquare), 10)
	g, _ := bigG.SetString(config.GetString(paillpublickeyG), 10)
	return hmEncryption.PaillierPublickey{
		N:       n,
		Nsquare: nsquare,
		G:       g,
	}
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
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	archiveDb, err := hyperdb.GetArchiveDbByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return hyperstate.New(common.BytesToHash(latestBlk.MerkleRoot), db, archiveDb, conf, height, namespace)
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
	state, err := hyperstate.New(common.BytesToHash(merkleRoot), db, nil, conf, height, namespace)
	return state, closer, err
}

func substr(str string, start int, end int) string {
	rs := []rune(str)

	return string(rs[start:end])
}

type IntervalArgs struct {
	From         *BlockNumber    `json:"from"`
	To           *BlockNumber    `json:"to"`
	ContractAddr *common.Address `json:"address"`
	MethodID     string          `json:"methodID"`
}

type IntervalTime struct {
	StartTime int64 `json:"startTime"`
	Endtime   int64 `json:"endTime"`
}

type intArgs struct {
	from uint64
	to   uint64
}

// prepareExcute checks if arguments are valid.
// 0 value for txType means sending normal transaction, 1 means deploying contract, 2 means invoking contract,
// 3 means signing hash, 4 means maintaining contract.
func prepareExcute(args SendTxArgs, txType int) (SendTxArgs, error) {
	if args.From.Hex() == (common.Address{}).Hex() {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address 'from' is invalid"}
	}
	if (txType == 0 || txType == 2 || txType == 4) && args.To == nil {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address 'to' is invalid"}
	}
	if args.Timestamp <= 0 || (5*int64(time.Minute)+time.Now().UnixNano()) < args.Timestamp {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'timestamp' is invalid"}
	}
	if txType != 3 && args.Signature == "" {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'signature' can't be empty"}
	}
	if args.Nonce <= 0 {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'nonce' is invalid"}
	}
	if txType == 4 && args.Opcode == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if txType == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if args.SnapshotId != "" && args.Simulate != true {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "can not query history ledger without `simulate` mode"}
	}
	if args.Timestamp+time.Duration(24*time.Hour).Nanoseconds() < time.Now().UnixNano() {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "transaction out of date"}
	}

	return args, nil
}

// prepareIntervalArgs checks if arguments are valid.
// If the client send BlockNumber "", it will be converted to 0. If client send BlockNumber 0, it will return error.
func prepareIntervalArgs(args IntervalArgs, namespace string) (*intArgs, error) {
	if args.From == nil || args.To == nil {
		return nil, &common.InvalidParamsError{Message: "missing params 'from' or 'to'"}
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
			return nil, &common.InvalidParamsError{Message: "invalid params from or to"}
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

// preparePagingArgs checks if paging arguments are valid.
func preparePagingArgs(args PagingArgs) (PagingArgs, error) {
	if args.PageSize == 0 {
		return PagingArgs{}, &common.InvalidParamsError{Message: "'pageSize' can't be zero or empty"}
	} else if args.Separated%args.PageSize != 0 {
		return PagingArgs{}, &common.InvalidParamsError{Message: "invalid 'pageSize' or 'separated'"}
	} else if args.MaxBlkNumber == BlockNumber(0) || args.MinBlkNumber == BlockNumber(0) {
		return PagingArgs{}, &common.InvalidParamsError{Message: "'minBlkNumber' or 'maxBlkNumber' can't be zero or empty"}
	} else if args.ContractAddr == nil {
		return PagingArgs{}, &common.InvalidParamsError{Message: "'address' can't be empty"}
	}

	return args, nil
}
