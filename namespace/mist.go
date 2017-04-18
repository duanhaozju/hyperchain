package namespace

import (
	ledger "hyperchain/core/vm/jcee/go/ledger"
)

type JvmManager struct {
	ledgerProxy      *ledger.LedgerProxy
}

func NewJvmManager() *JvmManager {
	return &JvmManager{
		ledgerProxy:     ledger.NewLedgerProxy(),
	}
}
