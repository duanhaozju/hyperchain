//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import "errors"

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

var (
	// ErrDuplicateTx is returned if the transaction is a duplicate one.
	ErrDuplicateBatch = errors.New("duplicate batch")

	ErrNoBatch = errors.New("can't find batch with id")

	ErrNoTxHash = errors.New("can't find tx hash in batched txs")

	ErrMismatch = errors.New("mismatch tx hash after receive return fetch txs")

	ErrMissing = errors.New("missing some txs compared with hashList from primary")

	ErrNoCachedBatch = errors.New("missing batch in cached batch list")

	ErrDuplicateTx = errors.New("duplicate transaction")

	ErrPoolFull = errors.New("txPool is full")

	ErrEmptyFull = errors.New("txPool is empty")
)
