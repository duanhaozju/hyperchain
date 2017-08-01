package api

import (
	"errors"
	"hyperchain/common"
	"hyperchain/core/hyperstate"
	"hyperchain/crypto/hmEncryption"
	"math/big"
	"time"
	edb "hyperchain/core/db_utils"
	"hyperchain/hyperdb"
	"hyperchain/core/vm"
)

const (
	rateLimitEnable = "flow.control.ratelimit.enable"

	TRANSACTION         = "transaction"
	transactionPeak     = "flow.control.ratelimit.txRatePeak"
	transactionFillRate = "flow.control.ratelimit.txFillRate"

	CONTRACT         = "contract"
	contractPeak     = "flow.control.ratelimit.contractRatePeak"
	contractFillRate = "flow.control.ratelimit.contractFillRate"

	paillpublickeyN       = "global.configs.hmpublickey.N"
	paillpublickeynsquare = "global.configs.hmpublickey.Nsquare"
	paillpublickeyG       = "global.configs.hmpublickey.G"
)

// getRateLimitEnable - get rate limit switch value
func getRateLimitEnable(conf *common.Config) bool {
	return conf.GetBool(rateLimitEnable)
}

// getRateLimitPeak - get rate limit peak value
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

// getFillRate - get rate limit fill speed
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

// getPaillierPublickey - get public key for hmEncryption
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
	archieveDb, err := hyperdb.GetArchieveDbByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return hyperstate.New(common.BytesToHash(latestBlk.MerkleRoot), db, archieveDb, conf, height, namespace)
}

func substr(str string, start int, end int) string {
	rs := []rune(str)

	return string(rs[start:end])
}