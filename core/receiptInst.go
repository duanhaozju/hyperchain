package core

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"sync"
)

type ReceiptMap map[common.Hash]*types.Receipt

type ReceiptInst struct {
	data      ReceiptMap   // store in db, synchronization with block
	lock      sync.RWMutex // the lock for receipt of reading of writing
	state     stateType    // the receipt cache state, use for singleton
	stateLock sync.Mutex   // the lock of get receipt cache instance
}

var receiptInst = &ReceiptInst{
	state: closed,
}

func GetReceiptInst() (*ReceiptInst, error) {
	receiptInst.stateLock.Lock()
	defer receiptInst.stateLock.Unlock()

	if receiptInst.state == closed {
		receiptInst.data = make(ReceiptMap)
		receiptInst.state = opened
	}
	return receiptInst, nil
}

func (self *ReceiptInst) PutReceipt(txHash common.Hash, value *types.Receipt) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.data[txHash] = value
}

func (self *ReceiptInst) GetReceipt(txHash common.Hash) *types.Receipt {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.data[txHash]
}

func (self *ReceiptInst) DeleteReceipt(txHash common.Hash) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.data, txHash)
}

func (self *ReceiptInst) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	for k := range self.data {
		delete(self.data, k)
	}
}

//-- GetAllCacheBalance get all cacheBalance
func (self *ReceiptInst) GetAllReceipt() ReceiptMap {
	self.lock.RLock()
	defer self.lock.RUnlock()
	var ret = make(ReceiptMap)
	for key, value := range self.data {
		ret[key] = value
	}
	return ret
}
func (self *ReceiptInst) Commit() error {
	self.lock.RLock()
	defer self.lock.RUnlock()
	receipts := make([]*types.Receipt, 0)
	for _, v := range self.data {
		receipts = append(receipts, v)
	}
	return WriteReceipts(receipts)
}
