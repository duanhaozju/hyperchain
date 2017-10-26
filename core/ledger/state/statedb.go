// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"bytes"
	"encoding/json"
	"github.com/deckarep/golang-set"
	"github.com/hyperchain/hyperchain/common"
	cm "github.com/hyperchain/hyperchain/core/common"
	"github.com/hyperchain/hyperchain/core/ledger/tree/bucket"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"sync/atomic"
)

// Revision one snapshot meta information.
type revision struct {
	id           int
	journalIndex int
}

// StateDBs within the hyperchain protocol are used to store anything
// * Contracts
// * Accounts
type StateDB struct {
	db        db.Database
	archiveDb db.Database
	root      common.Hash

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*StateObject
	stateObjectsDirty map[common.Address]struct{}

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash]types.Logs
	logSize      uint

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        Journal
	validRevisions []revision
	nextRevisionId int

	// bucket tree related
	conf *common.Config
	tree *bucket.BucketTree
	lock sync.Mutex
	// current block pool status
	curSeqNo    uint64 // current seqNo in process
	oldestSeqNo uint64 // oldest seqNo in content cache cache
	// atomic related
	batchCache   *common.Cache // use to store batch handler for different block process
	archiveCache *common.Cache // use to store batch handler for different block process
	contentCache *common.Cache // use to store modification set for different block process

	logger *logging.Logger
}

// New creates a new state with the given root and height.
func New(root common.Hash, db, archiveDb db.Database, conf *common.Config, height uint64) (*StateDB, error) {
	// initialize bucket tree
	bucketTree := bucket.NewBucketTree(db, string(compositeStateBucketPrefix()))
	// initialize cache
	batchCache, _ := common.NewCache()
	contentCache, _ := common.NewCache()
	archiveCache, _ := common.NewCache()
	state := &StateDB{
		db:                db,
		archiveDb:         archiveDb,
		root:              root,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		logs:              make(map[common.Hash]types.Logs),
		conf:              conf,
		tree:              bucketTree,
		batchCache:        batchCache,
		contentCache:      contentCache,
		archiveCache:      archiveCache,
		logger:            common.GetLogger(db.Namespace(), "state"),
	}
	err := bucketTree.Initialize(initTreeConfig(state.getCapacity(STATEDB), state.getAggreation(STATEDB),
		state.getMerkleCacheSize(STATEDB), state.getBucketCacheSize(STATEDB)))
	curHash, err := bucketTree.Process()
	state.logger.Debugf("latest state root %s", common.Bytes2Hex(curHash))
	if err != nil {
		state.logger.Errorf("new state db failed, error message %s", err.Error())
		return nil, err
	}
	if !validateRoot(root, common.BytesToHash(curHash)) {
		state.logger.Errorf("invalid root, correct root hash of state bucket tree is %s", common.Bytes2Hex(curHash))
		return nil, errors.New("invalid state bucket tree root")
	}
	// set oldest seqNo
	state.logger.Debugf("oldest height when initialize %d", height+1)
	state.setLatest(height + 1)
	return state, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the latest modified objects to avoid reloading data for the next operations.
func (self *StateDB) Reset() error {
	self.lock.Lock()
	defer self.lock.Unlock()
	// save modification set to content cache
	if self.curSeqNo == 0 {
		self.logger.Warning("make sure current seqNo is zero during the state reset")
	}
	// reset obj.onDirty callback function
	dirtyCopy := make(map[common.Address]*StateObject)
	for _, obj := range self.stateObjects {
		if _, dirty := self.stateObjectsDirty[obj.address]; dirty == true && !cm.IsPrecompiledAccount(obj.address) {
			obj.onDirty = self.MarkStateObjectDirty
			dirtyCopy[obj.address] = obj
		}
	}
	self.contentCache.Add(self.curSeqNo, dirtyCopy)
	if len(dirtyCopy) > 0 {
		self.logger.Debugf("save validation result to content cache, with %d element", len(dirtyCopy))
	}
	// clear all stuff
	self.stateObjects = make(map[common.Address]*StateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash]types.Logs)
	self.logSize = 0
	return nil
}

// MarkProcessStart marks a block's process is begin to initialize some stuff.
func (self *StateDB) MarkProcessStart(seqNo uint64) {
	// set current seqNo
	self.logger.Debugf("current process seqNo #%d", seqNo)
	atomic.StoreUint64(&self.curSeqNo, seqNo)
}

// MarkProcessFinish marks a block's process has finished.
// Remove cached content in cache to avoid of memory leak.
func (self *StateDB) MarkProcessFinish(seqNo uint64) {
	// remove all batch handler with related seqNo less than `seqNo`
	// because if empty valiation event happen(no validate transaction in a batch), this type of validation event
	// doesn't have commit event. Which can lead to memory leak if just remove seqNo related batch and content
	judge := func(key interface{}, iterKey interface{}) bool {
		id := key.(uint64)
		iterId := iterKey.(uint64)
		if id >= iterId {
			return true
		}
		return false
	}
	self.batchCache.RemoveWithCond(seqNo, judge)
	// remove content
	self.logger.Debugf("finish seqNo #%d processing, move oldest from #%d to #%d", seqNo, atomic.LoadUint64(&self.oldestSeqNo), seqNo+1)
	atomic.StoreUint64(&self.oldestSeqNo, seqNo+1)
	// remove all content with related seqNo less than `seqNo`
	self.contentCache.RemoveWithCond(seqNo, judge)
}

// setLatest sets oldest when state been initialize
// seqNo should be the chain's height.
func (self *StateDB) setLatest(seqNo uint64) {
	atomic.StoreUint64(&self.oldestSeqNo, seqNo)
}

// Purge clears out all uncommitted data.
func (self *StateDB) Purge() {
	self.batchCache.Purge()
	self.contentCache.Purge()
	self.archiveCache.Purge()
	self.tree.Clear()

	self.stateObjects = make(map[common.Address]*StateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash]types.Logs)
	self.logSize = 0
}

// ResetToTarget resets oldest seqNo and root to target.
func (self *StateDB) ResetToTarget(oldest uint64, root common.Hash) {
	atomic.StoreUint64(&self.oldestSeqNo, oldest)
	self.root = root
	self.Purge()
}

// FetchBatch fetchs a batch from batch collection with correspondent seqNo
// create a new batch if not exist in cache.
func (self *StateDB) FetchBatch(seqNo uint64, typ int) db.Batch {
	var c *common.Cache
	if typ == BATCH_NORMAL {
		c = self.batchCache
	} else {
		c = self.archiveCache
	}
	if c.Contains(seqNo) {
		// already exist
		self.logger.Debugf("fetch batch for #%d exist in batch cache", seqNo)
		batch, _ := c.Get(seqNo)
		return batch.(db.Batch)
	} else {
		// not exist right now
		batch := self.db.NewBatch()
		self.logger.Debugf("create one batch for #%d, %p", seqNo, batch)
		c.Add(seqNo, batch)
		return batch
	}
}

// MakeArchive flushs all cached archived content to historical database.
func (self *StateDB) MakeArchive(seqNo uint64) {
	batch := self.FetchBatch(seqNo, BATCH_ARCHIVE)
	self.archiveCache.Remove(seqNo)
	self.logger.Debugf("make archieve seqNo %d, totally %d elements contained.", seqNo, batch.Len())
	go batch.Write()
}

// DeleteBatch deletes a batch handler in cache with correspondent seqNo
// avoid of memory leak.
func (self *StateDB) DeleteBatch(seqNo uint64) {
	self.logger.Criticalf("remove batch for #%d from batch cache", seqNo)
	self.batchCache.Remove(seqNo)
}

// CompareRoot compares the current state root value with the input
func (self *StateDB) CompareRoot(root common.Hash) bool {
	return bytes.Compare(root.Bytes(), self.root.Bytes()) == 0
}

// StartRecord marks a transaction process's beginning.
func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

// AddLog adds logs generated in vm to state,
func (self *StateDB) AddLog(log *types.Log) {
	self.journal.JournalList = append(self.journal.JournalList, &AddLogChange{Txhash: self.thash})
	log.TxHash = self.thash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

// GetLogs obtains logs by transaction hash
func (self *StateDB) GetLogs(hash common.Hash) types.Logs {
	return self.logs[hash]
}

// Logs gets all logs in state
func (self *StateDB) Logs() types.Logs {
	var logs types.Logs
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existant
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.GetStateObject(addr)
	return so == nil || so.empty()
}

// GetAccount retrieves account by address.
func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	ret := self.GetStateObject(addr)
	if ret != nil {
		return ret
	} else {
		return nil
	}
}

// Get all accounts in database.
func (self *StateDB) GetAccounts() map[string]vm.Account {
	ret := make(map[string]vm.Account)
	iter := self.db.NewIterator([]byte(accountPrefix))
	for iter.Next() {
		addr, ok := splitCompositeAccountKey(iter.Key())
		if ok == false {
			continue
		}
		address := common.BytesToAddress(addr)
		var account Account
		json.Unmarshal(iter.Value(), &account)
		newobj := newObject(self, address, account, nil, false, nil, self.logger)
		ret[address.Hex()] = newobj
	}
	return ret
}

// GetBalance retrieves the balance from the given address or 0 if object not found.
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

// GetNonce nonce is the contract number state object has deployed.
func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

// GetCode code is the contract's code.
func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(self.db)
	}
	return nil
}

// GetCodeSize gets code len.
func (self *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return 0
	}
	return len(stateObject.Code(self.db))
}

// GetCodeHash gets code hash.
func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

// GetState retrieves specified contract state with given address and state key.
func (self *StateDB) GetState(a common.Address, b common.Hash) (bool, []byte) {
	// Prefer `live` object
	var liveObj *StateObject
	if obj := self.stateObjects[a]; obj != nil {
		liveObj = obj
		if obj.suicided {
			return false, nil
		} else {
			existed, value := obj.GetState(b)
			// if storage entry exist in live object's storage cache
			if existed {
				self.logger.Debugf("get state for %s in live objects, key %s, value %s", a.Hex(), b.Hex(), common.Bytes2Hex(value))
				return true, value
			}
		}
	}
	// find in content cache
	if self.curSeqNo > 0 {
		for i := self.curSeqNo - 1; i >= atomic.LoadUint64(&self.oldestSeqNo); i -= 1 {
			res, existed := self.contentCache.Get(i)
			if existed == false {
				self.logger.Error("load content from content cache failed")
				continue
			}
			content := res.(map[common.Address]*StateObject)
			if obj := content[a]; obj != nil {
				if obj.suicided {
					return false, nil
				} else {
					existed, value := obj.GetState(b)
					if existed {
						self.logger.Debugf("get state for %s in content cache, key %s, value %s", a.Hex(), b.Hex(), common.Bytes2Hex(value))
						if liveObj == nil {
							// save obj itself to current cache
							self.setStateObject(obj)
						} else {
							// save into live obj's cache storage avoid disk cost for next fetch
							liveObj.cachedStorage[b] = value
						}
						return true, value
					}
				}
			}
		}
	}
	// load from database
	existed, value := GetStateFromDB(self.db, a, b)
	if existed {
		// add related obj to live cache
		if liveObj == nil {
			self.logger.Debugf("get state for %s in database, key %s, value %s, add to live state object's storage cache", a.Hex(), b.Hex(), common.Bytes2Hex(value))
			// Load the object from the database.
			data, err := self.db.Get(compositeAccountKey(a.Bytes()))
			if err != nil {
				return false, nil
			}
			var account Account
			err = Unmarshal(data, &account)
			if err != nil {
				return false, nil
			}
			// Insert into the live set.
			obj := newObject(self, a, account, self.MarkStateObjectDirty, true, initTreeConfig(self.getCapacity(STATEOBJ), self.getAggreation(STATEOBJ), self.getMerkleCacheSize(STATEOBJ), self.getBucketCacheSize(STATEOBJ)), self.logger)
			obj.cachedStorage[b] = value
			self.setStateObject(obj)
		} else {
			// save into live obj's cache storage avoid disk cost for next fetch
			self.logger.Debugf("get state for %s in database, key %s, value %s, add %s to live objects", a.Hex(), b.Hex(), common.Bytes2Hex(value), a.Hex())
			liveObj.cachedStorage[b] = value
		}
		return true, value
	}
	self.logger.Debugf("find state for %s %s failed", a.Hex(), b.Hex())
	return false, nil
}

// GetDeployedContract retrieves deployed contract list.
func (self *StateDB) GetDeployedContract(addr common.Address) []string {
	obj := self.GetStateObject(addr)
	if obj != nil {
		return obj.DeployedContracts()
	}
	return nil
}

// AddDeployedContract adds a new created contract address to the maintain list.
func (self *StateDB) AddDeployedContract(addr common.Address, contract common.Address) {
	self.logger.Debugf("state object %s add contract %s to deployed list", addr.Hex(), contract.Hex())
	creator := self.GetStateObject(addr)
	if creator == nil {
		self.logger.Noticef("no state object %s found", addr.Hex())
		return
	}
	creator.AppendDeployedContract(contract)
}

// GetCreator retrieve creator address.
// If this account is an EOA(Enternal Owned Account), the creator is 0x0000000,
// Otherwise, for contract account, returns the actual creator.
func (self *StateDB) GetCreator(addr common.Address) common.Address {
	obj := self.GetStateObject(addr)
	if obj != nil {
		return obj.Creator()
	}
	return common.Address{}
}

// SetCreator sets creator.
func (self *StateDB) SetCreator(addr common.Address, creator common.Address) {
	self.logger.Debugf("state object %s set creator as %s", addr.Hex(), creator.Hex())
	obj := self.GetStateObject(addr)
	if obj == nil {
		self.logger.Noticef("no state object %s found", addr.Hex())
		return
	}
	obj.SetCreator(creator)
}

// GetStatus gets account status.
func (self *StateDB) GetStatus(address common.Address) int {
	obj := self.GetStateObject(address)
	if obj != nil {
		return obj.Status()
	}
	return -1
}

// SetStatus set specific account's status with given one.
func (self *StateDB) SetStatus(address common.Address, status int) {
	obj := self.GetStateObject(address)
	if obj == nil {
		self.logger.Noticef("no state object %s found", address.Hex())
		return
	}
	obj.SetStatus(status)
}

// GetCreateTime returns the brith block number of object.
func (self *StateDB) GetCreateTime(address common.Address) uint64 {
	obj := self.GetStateObject(address)
	if obj != nil {
		return obj.CreateTime()
	}
	return 0
}

// SetCreateTime sets the brith block number of object.
func (self *StateDB) SetCreateTime(address common.Address, time uint64) {
	obj := self.GetStateObject(address)
	if obj == nil {
		self.logger.Noticef("no state object %s found", address.Hex())
		return
	}
	obj.SetCreateTime(time)
}

// GetCurrentTxHash returns current transaction hash under processing.
func (self *StateDB) GetCurrentTxHash() common.Hash {
	return self.thash
}

// AddBalance adds balance to an account.
func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	} else {
		self.logger.Noticef("state object %s doesn't exist or has been suicide", addr.Hex())
	}
}

// SetBalance sets balance.
func (self *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	} else {
		self.logger.Noticef("state object %s doesn't exist or has been suicide", addr.Hex())
	}
}

// SetNonce sets nonce.
func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	} else {
		self.logger.Noticef("state object %s doesn't exist or has been suicide", addr.Hex())
	}
}

// SetCode sets code.
func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(common.BytesToHash(crypto.Keccak256(code)), code)
	} else {
		self.logger.Noticef("state object %s doesn't exist or has been suicide", addr.Hex())
	}
}

// SetState sets a storage entry to a state object.
func (self *StateDB) SetState(addr common.Address, key common.Hash, value []byte, opcode int32) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		self.logger.Debug("hyper statedb set state find state object in live objects")
		stateObject.SetState(self.db, key, value, opcode)
	} else {
		self.logger.Noticef("state object %s doesn't exist or has been suicide", addr.Hex())
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// GetStateObject will return a non-nil account after Suicide.
func (self *StateDB) Delete(addr common.Address) bool {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return false
	}
	self.journal.JournalList = append(self.journal.JournalList, &SuicideChange{
		Account:     &addr,
		Prev:        stateObject.suicided,
		Prevbalance: new(big.Int).Set(stateObject.Balance()),
		PreObject:   stateObject,
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)
	return true
}

//
// Setting, updating & deleting state object methods
//

// updateStateObject writes the given object to the database.
func (self *StateDB) updateStateObject(batch db.Batch, stateObject *StateObject) []byte {
	addr := stateObject.Address()
	data, err := stateObject.Marshal()
	if err != nil {
		self.logger.Error("marshal stateobject failed", addr.Hex())
	}
	batch.Put(compositeAccountKey(addr.Bytes()), data)
	return data
}

// deleteStateObject removes the given object from the database
func (self *StateDB) deleteStateObject(batch db.Batch, stateObject *StateObject) {
	self.logger.Debugf("delete state object %s during state commit, seqNo #%d", stateObject.address.Hex(), self.curSeqNo)
	stateObject.deleted = true
	addr := stateObject.Address()
	batch.Delete(compositeAccountKey(addr.Bytes()))
	// delete related storage content
	iter := self.db.NewIterator(getStorageKeyPrefix(stateObject.address.Bytes()))
	for iter.Next() {
		batch.Delete(iter.Key())
	}
}

// GetStateObject retrieves a state object given my the address. Returns nil if not found.
func (self *StateDB) GetStateObject(addr common.Address) *StateObject {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.suicided {
			self.logger.Debugf("search state object %x in the live objects, but it has been suicide", addr)
			return nil
		}
		self.logger.Debugf("search state object %x in the live objects", addr)
		return obj
	}
	// Load from the content cache
	if self.curSeqNo != 0 {
		for i := self.curSeqNo - 1; i >= atomic.LoadUint64(&self.oldestSeqNo); i -= 1 {
			res, existed := self.contentCache.Get(i)
			if existed == false {
				self.logger.Error("load content from content cache failed")
				continue
			}
			content := res.(map[common.Address]*StateObject)
			if obj := content[addr]; obj != nil {
				if obj.suicided {
					self.logger.Noticef("search state object %x in the content cache, but it has been suicide", addr)
					return nil
				}
				self.logger.Debugf("search state object %x in the content cache, add it to live objects", addr)
				self.logger.Debugf("obj found from cached collection. totally with %d elements", len(obj.cachedStorage))
				self.setStateObject(obj)
				return obj
			}
		}
	}
	// Load the object from the database.
	data, err := self.db.Get(compositeAccountKey(addr.Bytes()))
	if err != nil {
		self.logger.Debugf("no state object been find")
		return nil
	}
	var account Account
	err = Unmarshal(data, &account)
	if err != nil {
		return nil
	}
	// Insert into the live set.
	self.logger.Debugf("find state object %x in database, add it to live objects", addr)
	obj := newObject(self, addr, account, self.MarkStateObjectDirty, true, initTreeConfig(self.getCapacity(STATEOBJ), self.getAggreation(STATEOBJ), self.getMerkleCacheSize(STATEOBJ), self.getBucketCacheSize(STATEOBJ)), self.logger)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *StateObject) {
	self.stateObjects[object.Address()] = object
}

// MarkStateObjectDirty adds the specified object to the dirty map to avoid costly
// state object cache iteration to find a handful of modified ones.
func (self *StateDB) MarkStateObjectDirty(addr common.Address) {
	self.stateObjectsDirty[addr] = struct{}{}
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (self *StateDB) createObject(addr common.Address) (newobj, prev *StateObject) {
	prev = self.GetStateObject(addr)
	newobj = newObject(self, addr, Account{}, self.MarkStateObjectDirty, true, initTreeConfig(self.getCapacity(STATEOBJ), self.getAggreation(STATEOBJ), self.getMerkleCacheSize(STATEOBJ), self.getBucketCacheSize(STATEOBJ)), self.logger)
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		self.logger.Infof("(+) %x\n", addr)
		self.journal.JournalList = append(self.journal.JournalList, &CreateObjectChange{Account: &addr})
	} else {
		self.journal.JournalList = append(self.journal.JournalList, &ResetObjectChange{Prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (self *StateDB) CreateAccount(addr common.Address) vm.Account {
	new, prev := self.createObject(addr)
	if prev != nil {
		new.setBalance(prev.data.Balance)
	}
	return new
}

// Snapshot returns an identifier for the current revision of the state.
func (self *StateDB) Snapshot() interface{} {
	id := self.nextRevisionId
	self.nextRevisionId++
	self.validRevisions = append(self.validRevisions, revision{id, len(self.journal.JournalList)})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(copy interface{}) {
	// Find the snapshot in the stack of valid snapshots.
	revid, ok := copy.(int)
	if ok == false {
		self.logger.Error("Revert to snapshot", copy, "failed")
		return
	}
	idx := sort.Search(len(self.validRevisions), func(i int) bool {
		return self.validRevisions[i].id >= revid
	})
	if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes.
	for i := len(self.journal.JournalList) - 1; i >= snapshot; i-- {
		// undo in memory
		// parameters *stateDB, batch, writeThrough, flush, sync
		self.journal.JournalList[i].Undo(self, nil, nil, false)
		self.logger.Infof("undo operation: %s", self.journal.JournalList[i])
	}
	self.journal.JournalList = self.journal.JournalList[:snapshot]

	// Remove invalidated snapshots from the stack.
	self.validRevisions = self.validRevisions[:idx]
}

// RevertToJournal reverts all state changes made since the target height.
func (self *StateDB) RevertToJournal(target uint64, current uint64, targetRoot []byte, batch db.Batch) error {
	var (
		dirtyObjects      = mapset.NewSet()
		dirtyObjectHashes = make(map[common.Address][]byte)
	)

	journalCache := NewJournalCache(self.db, self.logger)
	for i := current; i >= target+1; i -= 1 {
		self.logger.Debugf("undo changes for #%d", i)
		j, err := self.db.Get(compositeJournalKey(uint64(i)))
		if err != nil {
			self.logger.Debugf("get journal in database for #%d failed. make sure #%d doesn't have state change",
				i, i)
			continue
		}
		journal, err := UnmarshalJournal(j)
		if err != nil {
			self.logger.Errorf("unmarshal journal for #%d failed", i)
			continue
		}
		// undo journal in reverse
		for j := len(journal.JournalList) - 1; j >= 0; j -= 1 {
			self.logger.Debugf("journal %s", journal.JournalList[j].String())
			journal.JournalList[j].Undo(self, journalCache, batch, true)
			if journal.JournalList[j].GetType() == StorageHashChangeType {
				tmp := journal.JournalList[j].(*StorageHashChange)
				dirtyObjects.Add(*tmp.Account)
				dirtyObjectHashes[*tmp.Account] = tmp.Prev
			}
		}
		// remove persisted journals
		batch.Delete(compositeJournalKey(uint64(i)))
	}

	if err := journalCache.Flush(batch); err != nil {
		self.logger.Errorf("flush modified content failed. %s", err.Error())
		return err
	}
	// revert related stateObject storage bucket tree
	for addr := range dirtyObjects.Iter() {
		address := addr.(common.Address)
		bucketTree := bucket.NewBucketTree(self.db, string(compositeStorageBucketPrefix(address.Hex())))
		bucketTree.Initialize(initTreeConfig(self.getCapacity(STATEOBJ), self.getAggreation(STATEOBJ),
			self.getMerkleCacheSize(STATEOBJ), self.getBucketCacheSize(STATEOBJ)))
		// clear out cached stuff in global cache.
		bucketTree.Clear()
		// don't flush into disk util all operations finish
		bucketTree.Prepare(journalCache.getWorkingSet(ObjectWorkingSet, address))
		hash, _ := bucketTree.Process()
		bucketTree.Commit(batch)
		self.logger.Debugf("re-compute %s storage hash %s", address.Hex(), common.Bytes2Hex(hash))
		expect := dirtyObjectHashes[address]
		if common.BytesToHash(hash).Hex() != common.BytesToHash(expect).Hex() {
			self.logger.Errorf("after revert to #%d, state object %s revert failed, required storage hash %s, got %s",
				target, address.Hex(), common.Bytes2Hex(expect), common.Bytes2Hex(hash))
			return errors.New("revert state failed.")
		}
	}
	// revert state bucket tree
	tree := self.tree
	tree.Initialize(initTreeConfig(self.getCapacity(STATEDB), self.getAggreation(STATEDB),
		self.getMerkleCacheSize(STATEDB), self.getBucketCacheSize(STATEDB)))
	tree.Clear()
	tree.Prepare(journalCache.getWorkingSet(StateWorkingSet, common.Address{}))
	curHash, err := tree.Process()
	if err != nil {
		self.logger.Errorf("re-compute state bucket tree hash failed, error :%s", err.Error())
		return err
	}
	tree.Commit(batch)
	self.logger.Debugf("re-compute state hash %s", common.Bytes2Hex(curHash))
	if bytes.Compare(common.BytesToHash(curHash).Bytes(), targetRoot) != 0 {
		self.logger.Errorf("revert to a different state, required %s, but current state %s",
			common.Bytes2Hex(targetRoot), common.BytesToHash(curHash).Hex())
		return errors.New("revert state failed")
	}
	// revert state instance oldest and root
	self.ResetToTarget(uint64(target+1), common.BytesToHash(targetRoot))
	self.logger.Debugf("revert state from #%d to #%d success", current, target)
	return nil

}

// Commit commits all state changes to the database.
func (s *StateDB) Commit() (common.Hash, error) {
	root, _ := s.CommitBatch(deleteEmptyObjects)
	return root, nil
}

// CommitBatch commits all state changes to a write batch but does not
// execute the batch. It is used to validate state changes against
// the root hash stored in a block.
func (s *StateDB) CommitBatch(deleteEmptyObjects bool) (root common.Hash, batch db.Batch) {
	curSeqNo := atomic.LoadUint64(&s.curSeqNo)
	batch = s.FetchBatch(curSeqNo, BATCH_NORMAL)
	archiveBatch := s.FetchBatch(curSeqNo, BATCH_ARCHIVE)
	root, _ = s.commit(batch, archiveBatch, deleteEmptyObjects)
	return root, batch
}

func (s *StateDB) clearJournalAndRefund() {
	// 1. persist current journal to disk
	batch := s.db.NewBatch()
	curSeqNo := atomic.LoadUint64(&s.curSeqNo)
	if len(s.journal.JournalList) != 0 {
		if curSeqNo == 0 {
			s.logger.Warningf("make sure the seqNo of journal is [%d]", curSeqNo)
		}
		journal, err := s.journal.Marshal()
		if err != nil {
			s.logger.Errorf("marshal seqNo [%d] journal failed", curSeqNo)
		}
		batch.Put(compositeJournalKey(curSeqNo), journal)
	}
	// 2. clear journal
	s.journal = Journal{}
	s.validRevisions = s.validRevisions[:0]
	batch.Write()
}

func (s *StateDB) commit(dbw db.Batch, archieveDb db.Batch, deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund()
	var set ChangeSet
	workingSet := bucket.NewEntries()
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			s.logger.Debugf("seqNo #%d, state object %s been suicide or clearing out for empty", s.curSeqNo, stateObject.address.Hex())
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(dbw, stateObject)
			if enableFakeHashFn {
				d := stateObject.address.Bytes()
				set = append(set, d)
			} else {
				workingSet[stateObject.address.Hex()] = nil
			}
		case isDirty:
			// Write any contract code associated with the state object
			s.logger.Debugf("seqNo #%d, state object %s been updated", s.curSeqNo, stateObject.address.Hex())
			if stateObject.code != nil && stateObject.dirtyCode {
				if err := dbw.Put(compositeCodeHash(stateObject.Address().Bytes(), stateObject.CodeHash()), stateObject.code); err != nil {
					return common.Hash{}, err
				}
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.Flush(dbw, archieveDb); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			d := s.updateStateObject(dbw, stateObject)
			// Add to change set
			if enableFakeHashFn {
				c := append(stateObject.address.Bytes(), d...)
				set = append(set, c)
			} else {
				d, _ := stateObject.MarshalJSON()
				workingSet[stateObject.address.Hex()] = d
			}
		}
	}
	// Write trie changes.
	// Calculate Hash
	if enableFakeHashFn {
		s.root = SimpleHashFn(s.root, set)
	} else {
		// Use bucket tree instead
		s.logger.Debugf("begin to calculate state db root hash for #%d", s.curSeqNo)
		s.tree.Prepare(workingSet)
		hash, err := s.tree.Process()
		if err != nil {
			s.logger.Error("calculate hash for statedb failed")
			return common.Hash{}, err
		}
		s.root = common.BytesToHash(hash)
		s.tree.Commit(dbw)
	}
	return s.root, err
}

// validateRoot compares two root hash is equal or not.
func validateRoot(root common.Hash, curRoot common.Hash) bool {
	if root != curRoot {
		return false
	} else {
		return true
	}
}

// RecomputeCryptoHash rehashs the whole world state.
func (s *StateDB) RecomputeCryptoHash() (common.Hash, error) {
	var (
		dirtyObjects = bucket.NewEntries()
		accountIter  = s.db.NewIterator([]byte(accountPrefix))
	)
	defer accountIter.Release()
	for accountIter.Next() {
		addr, ok := splitCompositeAccountKey(accountIter.Key())
		if ok == false {
			continue
		}
		address := common.BytesToAddress(addr)
		var account Account
		err := json.Unmarshal(accountIter.Value(), &account)
		if err != nil {
			return common.Hash{}, err
		}
		stateObject := newObject(s, address, account, s.MarkStateObjectDirty, true, initTreeConfig(s.getCapacity(STATEOBJ), s.getAggreation(STATEOBJ), s.getMerkleCacheSize(STATEOBJ), s.getBucketCacheSize(STATEOBJ)), s.logger)
		dirtyStorage := bucket.NewEntries()
		iter := stateObject.db.db.NewIterator(getStorageKeyPrefix(stateObject.address.Bytes()))
		for iter.Next() {
			key, ok := splitCompositeStorageKey(stateObject.address.Bytes(), iter.Key())
			if ok == false {
				continue
			}
			valueCopy := make([]byte, len(iter.Value()))
			copy(valueCopy, iter.Value())

			if len(valueCopy) <= common.HashLength {
				valueCopy = common.BytesToHash(valueCopy).Bytes()
			}
			dirtyStorage[common.BytesToHash(key).Hex()] = valueCopy
		}
		stateObject.tree.Prepare(dirtyStorage)
		hash, err := stateObject.tree.Process()
		if err != nil {
			return common.Hash{}, err
		}
		stateObject.SetRoot(common.BytesToHash(hash))
		valueOfStateObject, _ := stateObject.MarshalJSON()
		dirtyObjects[stateObject.address.Hex()] = valueOfStateObject
		iter.Release()
	}
	s.tree.Prepare(dirtyObjects)
	hash, err := s.tree.Process()
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(hash), nil
}

// Merge merges another world status with current.
// Note, during the merge, account meta data should be merged first.
func (s *StateDB) Merge(db db.Database, batch db.Batch, expect common.Hash) error {
	applyAccount := func(k, v []byte) {
		blob, success := splitCompositeAccountKey(k)
		if !success {
			return
		}
		var account Account
		err := Unmarshal(v, &account)
		if err != nil {
			return
		}
		addr := common.BytesToAddress(blob)
		obj := newObject(s, addr, account, s.MarkStateObjectDirty, true, initTreeConfig(s.getCapacity(STATEOBJ), s.getAggreation(STATEOBJ), s.getMerkleCacheSize(STATEOBJ), s.getBucketCacheSize(STATEOBJ)), s.logger)
		s.setStateObject(obj)
		obj.onDirty(addr)
	}
	applyStorage := func(k, v []byte) {
		blob, success := retrieveAddrFromStorageKey(k)
		if !success {
			return
		}
		addr := common.BytesToAddress(blob)
		blob, success = splitCompositeStorageKey(addr.Bytes(), k)
		if !success {
			return
		}
		key := common.BytesToHash(blob)
		val := make([]byte, len(v))
		copy(val, v)
		s.SetState(addr, key, val, 0)
	}
	applyCode := func(k, v []byte) {
		blob, success := retrieveAddrFromCodeHash(k)
		if !success {
			return
		}
		addr := common.BytesToAddress(blob)
		blob, success = splitCompositeCodeHash(addr.Bytes(), k)
		if !success {
			return
		}
		codeHash := blob
		obj := s.GetStateObject(addr)
		if obj == nil {
			return
		}
		if bytes.Compare(codeHash, obj.CodeHash()) != 0 {
			return
		}
		s.SetCode(addr, v)
	}

	entryPrefixes := cm.RetrieveSnapshotFileds()
	for _, entry := range entryPrefixes {
		switch entry {
		case "-account":
			iter := db.NewIterator([]byte(entry))
			for iter.Next() {
				applyAccount(iter.Key(), iter.Value())
			}
			iter.Release()
		case "-storage":
			iter := db.NewIterator([]byte(entry))
			for iter.Next() {
				applyStorage(iter.Key(), iter.Value())
			}
			iter.Release()
		case "-code":
			iter := db.NewIterator([]byte(entry))
			for iter.Next() {
				applyCode(iter.Key(), iter.Value())
			}
			iter.Release()
		}
	}

	h, err := s.commit(batch, batch, deleteEmptyObjects)
	if err != nil {
		return err
	}
	if h != expect {
		return errors.New("different hash compared with recorded")
	}
	return nil
}

// NewIterator returns a iterator to traverse contract storage content.
func (self *StateDB) NewIterator(addr common.Address, slice *vm.IterRange) (vm.Iterator, error) {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		self.logger.Debug("hyper statedb set state find state object in live objects")
		var iter vm.Iterator
		if slice == nil {
			iter = NewStorageIterator(stateObject, nil, nil)
		} else {
			iter = NewStorageIterator(stateObject, slice.Start, slice.Limit)
		}
		return iter, nil
	} else {
		self.logger.Warningf("state object %s doesn't exist or has been suicide", addr.Hex())
		return nil, nil
	}
}
