// fetcher implements block operate
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package core

import (
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
	"hyperchain/core/types"
	"errors"

	"hyperchain/common"
)

type Fetcher struct {
	queues map[string]int          // Per peer block counts to prevent memory exhaustion

	queued map[common.Hash]*inject // Set of already queued blocks (to dedup imports)

	inject chan *inject
	queue  *prque.Prque
	quit   chan struct{}
}

var (
	errTerminated = errors.New("terminated")
)

type inject struct {
	block *types.Block
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

	for {

		for !f.queue.Empty() {

			op := f.queue.PopItem().(*inject)

			// If too high up the chain or phase, continue later

			f.insert(op.block)
		}

		select {
		case op := <-f.inject:

			f.enqueue( op.block)}
	}

}

//insert into db
func (f *Fetcher) insert(block *types.Block) {

	//TODO

}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *Fetcher) Enqueue(block *types.Block) error {
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

//put block into queue
func (f *Fetcher) enqueue(block *types.Block) {
	hash := block.BlockHash


	if _, ok := f.queued[common.BytesToHash(hash)]; !ok {
		op := &inject{

			block:  block,
		}

		f.queued[common.BytesToHash(hash)] = op
		f.queue.Push(op, -float32(block.Number))

	}
}
