package namespace

import (
	ledger "hyperchain/core/vm/jcee/go/ledger"
	"hyperchain/common"
)

type JvmManager struct {
	ledgerProxy      *ledger.LedgerProxy
}

func NewJvmManager(conf *common.Config) *JvmManager {
	return &JvmManager{
		ledgerProxy:     ledger.NewLedgerProxy(conf),
	}
}
