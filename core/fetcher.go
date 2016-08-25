package core

import (

	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
	"hyperchain-alpha/core/types"

	"hyperchain-alpha/common"
	"errors"

)

type Fetcher struct {

	queues map[string]int          // Per peer block counts to prevent memory exhaustion

	queued map[common.Hash]*inject // Set of already queued blocks (to dedup imports)

	inject chan *inject
	queue  *prque.Prque
	quit chan struct{}

}

var (
	errTerminated = errors.New("terminated")
)

type inject struct {

	block  *types.Block
}

func NewFetcher() *Fetcher {


	return &Fetcher{

		inject:         make(chan *inject),

		quit:           make(chan struct{}),

		queue:          prque.New(),
		queues:         make(map[string]int),
		queued:         make(map[common.Hash]*inject),

	}
}

func (f *Fetcher) Start() {





	for{

		for !f.queue.Empty() {

			op := f.queue.PopItem().(*inject)

			// If too high up the chain or phase, continue later

			f.insert( op.block)
		}
	}

}

func (f *Fetcher) insert( block *types.Block) {
	//TODO

}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *Fetcher) Enqueue( block *types.Block) error {
	op := &inject{

		block:  block,
	}
	select {
	case f.inject <- op:
		return nil
	case <-f.quit:
		return errTerminated
	}
}
// enqueue schedules a new future import operation, if the block to be imported
// has not yet been seen.
func (f *Fetcher) enqueue(peer string, block *types.Block) {
	hash := block.BlockHash


	count := f.queues[peer] + 1
	// Schedule the block for future importing
	if _, ok := f.queued[hash]; !ok {
		op := &inject{

			block:  block,
		}
		f.queues[peer] = count
		f.queued[hash] = op
		f.queue.Push(op, -float32(block.Number.Uint64()))


	}
}
