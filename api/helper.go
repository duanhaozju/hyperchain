package hpc

import (
	"errors"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/hyperstate"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/crypto/hmEncryption"
	"hyperchain/hyperdb"
	"math/big"
	"time"
)

const (
	stateType       = "global.structure.state"
	rateLimitEnable = "global.configs.ratelimit.enable"

	TRANSACTION         = "transaction"
	transactionPeak     = "global.configs.ratelimit.txRatePeak"
	transactionFillRate = "global.configs.ratelimit.txFillRate"

	CONTRACT         = "contract"
	contractPeak     = "global.configs.ratelimit.contractRatePeak"
	contractFillRate = "global.configs.ratelimit.contractFillRate"

	paillpublickeyN       = "global.configs.hmpublickey.N"
	paillpublickeynsquare = "global.configs.hmpublickey.Nsquare"
	paillpublickeyG       = "global.configs.hmpublickey.G"
)

// getStateType - get state type
func getStateType(conf *common.Config) string {
	return conf.GetString(stateType)
}

// getRateLimitEnable - get rate limit switch value
func getRateLimitEnable(conf *common.Config) bool {
	return conf.GetBool(rateLimitEnable)
}

// getRateLimitPeak - get rate limit peak value
func getRateLimitPeak(conf *common.Config, choice string) int64 {
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
func getFillRate(conf *common.Config, choice string) (time.Duration, error) {
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

func GetStateInstance(root common.Hash, db hyperdb.Database, conf *common.Config) (vm.Database, error) {
	switch getStateType(conf) {
	case "rawstate":
		return state.New(root, db)
	case "hyperstate":
		height := core.GetHeightOfChain()
		return hyperstate.New(root, db, conf, height)
	default:
		return nil, errors.New("no state type specified")
	}
}