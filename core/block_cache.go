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
	ValidTxs    []*types.Transaction
	SeqNo       uint64
}

type Data map[string]BlockRecord

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

func (self *BlockCache) Record(hash string, record BlockRecord) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.data[hash] = record
}

func (self *BlockCache) Get(hash string) BlockRecord {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.data[hash]
}

func (self *BlockCache) Delete(hash string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	seqNo := self.data[hash].SeqNo
	delete(self.data, hash)
	for k, v := range self.data {
		if v.SeqNo <= seqNo {
			delete(self.data, k)
		}
	}
}

func (self *BlockCache) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	for k := range self.data {
		delete(self.data, k)
	}
}

func (self *BlockCache) All() map[string]BlockRecord {
	self.lock.Lock()
	defer self.lock.Unlock()
	ret := make(map[string]BlockRecord)
	//var ret map[string]BlockRecord
	if self.data!=nil{
	for k, v := range self.data {
		ret[k] = v
	}
	}
	return ret
}
