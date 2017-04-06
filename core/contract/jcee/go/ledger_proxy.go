//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"github.com/pkg/errors"
	"hyperchain/core/contract/jcee/protos"
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
}

func (lp *LedgerProxy) ProcessCommand(cmd *contract.Command) ([]byte, error) {

	//TODO: parse cmd and execute it
	return nil, errors.New("not implement yet")
}
