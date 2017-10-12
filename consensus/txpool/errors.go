//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import "errors"

var (
	ErrNoBatch = errors.New("can't find batch with id")

	ErrNoTxHash = errors.New("can't find tx hash in batched txs")

	ErrMismatch = errors.New("Unmatch tx hash after receive return fetch txs")

	ErrDuplicateTx = errors.New("duplicate transaction")

	ErrPoolFull = errors.New("txPool is full")

	ErrPoolEmpty = errors.New("txPool is empty")
)
