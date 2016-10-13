package core

import (
	"hyperchain/core/types"
	"sync"
)

type BlockRecord struct {
	TxRoot      []byte
	ReceiptRoot []byte
	MerkleRoot  []byte
	InvalidTxs  []*types.InvalidTransactionRecord
}

type Data map[uint64]BlockRecord

type BlockCache struct {
	data      Data         // store in db, synchronization with block
	lock      sync.RWMutex // the lock for receipt of reading of writing
	state     stateType    // the receipt cache state, use for singleton
	stateLock sync.Mutex   // the lock of get receipt cache instance
}

var blockCache = &BlockCache{
	state: closed,
}

func GetBlockCache() (*BlockCache, error) {
	blockCache.stateLock.Lock()
	defer blockCache.stateLock.Unlock()

	if blockCache.state == closed {
		blockCache.data = make(Data)
		blockCache.state = opened
	}
	return blockCache, nil
}

func (self *BlockCache) Record(seqNo uint64, record BlockRecord) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.data[seqNo] = record
}

func (self *BlockCache) Get(seqNo uint64) BlockRecord {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.data[seqNo]
}

func (self *BlockCache) Delete(seqNo uint64) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.data, seqNo)
}

func (self *BlockCache) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	for k := range self.data {
		delete(self.data, k)
	}
}
