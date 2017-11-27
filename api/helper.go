package api

import (
	"errors"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/hmEncryption"
	"math/big"
	"time"
)

const (
	rateLimitEnable = "flow.control.ratelimit.enable"

	BLOCK               = "block"
	TRANSACTION         = "transaction"
	transactionPeak     = "flow.control.ratelimit.txRatePeak"
	transactionFillRate = "flow.control.ratelimit.txFillRate"

	contractPeak     = "flow.control.ratelimit.contractRatePeak"
	contractFillRate = "flow.control.ratelimit.contractFillRate"

	paillpublickeyN       = "global.configs.hmpublickey.N"
	paillpublickeynsquare = "global.configs.hmpublickey.Nsquare"
	paillpublickeyG       = "global.configs.hmpublickey.G"

	CONTRACT                = "contract"
)



// getRateLimitEnable returns rate limit switch value.
func GetRateLimitEnable(conf *common.Config) bool {
	return conf.GetBool(rateLimitEnable)
}

// getRateLimitPeak returns rate limit peak value.
func GetRateLimitPeak(namespace string, conf *common.Config, choice string) int64 {
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
func GetFillRate(namespace string, conf *common.Config, choice string) (time.Duration, error) {
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
func GetPaillierPublickey(config *common.Config) hmEncryption.PaillierPublickey {
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
