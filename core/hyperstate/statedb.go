package hyperstate

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"hyperchain/core/vm"
	"encoding/json"
	"hyperchain/tree/bucket"
	"github.com/pkg/errors"
	"sync/atomic"
)


const (
	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000
	// whether turn on fake hash function
	enableFakeHashFn = false
	// whether to remove empty stateObject
	deleteEmptyObjects = true
)

type revision struct {
	id           int
	journalIndex int
}

// StateDBs within the hyperchain protocol are used to store anything
// * Contracts
// * Accounts
type StateDB struct {
	db            hyperdb.Database
	root          common.Hash
	codeSizeCache *lru.Cache

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*StateObject
	stateObjectsDirty map[common.Address]struct{}

	// The refund counter, also used by state transitioning.
	refund *big.Int

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash]vm.Logs
	logSize      uint

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        Journal
	validRevisions []revision
	nextRevisionId int

	// bucket tree related
	bktConf        bucket.Conf
	bucketTree     *bucket.BucketTree
	lock           sync.Mutex
	// current block pool status
	curSeqNo       uint64        // current seqNo in process
	oldestSeqNo    uint64        // oldest seqNo in content cache cache
	// atomic related
	batchCache     *common.Cache // use to store batch handler for different block process
	contentCache   *common.Cache // use to store modification set for different block process
}

// Create a new state from a given root
func New(root common.Hash, db hyperdb.Database, bktConf bucket.Conf, height uint64) (*StateDB, error) {
	csc, _ := lru.New(codeSizeCacheSize)
	// initialize bucket tree
	bucketPrefix, _ := CompositeStateBucketPrefix()
	bucketTree := bucket.NewBucketTree(string(bucketPrefix))
	bucketTree.Initialize(SetupBucketConfig(bktConf.StateSize, bktConf.StateLevelGroup))
	// check root hash validation
	// initialize cache
	batchCache, _ := common.NewCache()
	contentCache, _ := common.NewCache()
	curHash, err := bucketTree.ComputeCryptoHash()
	log.Criticalf("latest state root %x", curHash)
	if err != nil {
		log.Errorf("new state db failed, error message %s", err.Error())
		return nil, err
	}
	if !validateRoot(root, common.BytesToHash(curHash)) {
		log.Errorf("invalid root, correct root hash of state bucket tree is %s", common.Bytes2Hex(curHash))
		return nil, errors.New("invalid state bucket tree root")
	}
	state := &StateDB{
		db:                db,
		root:              root,
		codeSizeCache:     csc,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]vm.Logs),
		bktConf:           bktConf,
		bucketTree:        bucketTree,
		batchCache:        batchCache,
		contentCache:      contentCache,
	}
	// set oldest seqNo
	log.Criticalf("oldest height when initialize %d", height + 1)
	state.LoadLatest(height + 1)
	return state, nil
}

// New creates a new statedb by reusing journalled data to avoid costly
// disk io.
// Deprecated
func (self *StateDB) New(root common.Hash) (*StateDB, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	bucketPrefix, _ := CompositeStateBucketPrefix()
	bucketTree := bucket.NewBucketTree(string(bucketPrefix))
	bucketTree.Initialize(SetupBucketConfig(self.bktConf.StateSize, self.bktConf.StateLevelGroup))

	return &StateDB{
		db:                self.db,
		codeSizeCache:     self.codeSizeCache,
		root:              root,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]vm.Logs),
		bktConf:           self.bktConf,
		bucketTree:        bucketTree,
	}, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset() error {
	log.Critical("reset state db")
	self.lock.Lock()
	defer self.lock.Unlock()
	// save modification set to content cache
	if self.curSeqNo == 0 {
		log.Warning("make sure current seqNo is zero during the state reset")
	}
	if len(self.stateObjects) > 0 {
		log.Criticalf("save validation result to content cache, with %d element", len(self.stateObjects))
	}
	// IMPORTANT reset obj.onDirty callback function and bucket tree
	for _, obj := range self.stateObjects {
		obj.onDirty = self.MarkStateObjectDirty
		obj.bucketTree.Reset()
	}
	self.contentCache.Add(self.curSeqNo, self.stateObjects)
	// clear all stuff
	self.stateObjects = make(map[common.Address]*StateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash]vm.Logs)
	self.logSize = 0
	// reset bucket tree
	self.bucketTree.Reset()
	return nil
}


// mark a block's process is begin
// initialize some stuff
func (self *StateDB) MarkProcessStart(seqNo uint64) {
	// reset state
	self.Reset()
	// set current seqNo
	log.Criticalf("current process seqNo #%d", seqNo)
	atomic.StoreUint64(&self.curSeqNo, seqNo)
}
// mark a block's process has finished
// remove some stuff in cache to avoid of memory leak
func (self *StateDB) MarkProcessFinish(seqNo uint64) {
	// remove all batch handler with related seqNo less than `seqNo`
	// cause if empty valiation event happen(no validate transaction in a batch), this type of validation event
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
	log.Criticalf("finish seqNo #%d processing, move oldest from #%d to #%d", seqNo, atomic.LoadUint64(&self.oldestSeqNo), seqNo+1)
	atomic.StoreUint64(&self.oldestSeqNo, seqNo + 1)
	// remove all content with related seqNo less than `seqNo`
	self.contentCache.RemoveWithCond(seqNo, judge)
}
// set oldest when state been initialize
// seqNo should be the chain's height
func (self *StateDB) LoadLatest(seqNo uint64) {
	atomic.StoreUint64(&self.oldestSeqNo, seqNo)
}

// batch cache related
// fetch a batch from batch cache with correspondent seqNo
// create a new batch if not exist in cache
func (self *StateDB) FetchBatch(seqNo uint64) hyperdb.Batch {
	if self.batchCache.Contains(seqNo) {
		// already exist
		log.Criticalf("fetch batch for #%d exist in batch cache", seqNo)
		batch, _ := self.batchCache.Get(seqNo)
		return batch.(hyperdb.Batch)
	} else {
		// not exist right now
		log.Criticalf("create one batch for #%d", seqNo)
		batch := self.db.NewBatch()
		self.batchCache.Add(seqNo, batch)
		return batch
	}
}
// delete a batch handler in cache with correspondent seqNo
// avoid memory leak
func (self *StateDB) DeleteBatch(seqNo uint64) {
	log.Criticalf("remove batch for #%d from batch cache", seqNo)
	self.batchCache.Remove(seqNo)
}
// mark a transaction process's beginning
func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}
// add logs generated in vm to state
// doesn't assign block hash now
// because the blcok hash hasn't been calculated
// correctly block  hash will be assigned in the commit phase
func (self *StateDB) AddLog(log *vm.Log) {
	self.journal.JournalList = append(self.journal.JournalList, &AddLogChange{Txhash: self.thash})
	log.TxHash = self.thash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}
// obtain logs by transaction hash
func (self *StateDB) GetLogs(hash common.Hash) vm.Logs {
	return self.logs[hash]
}
// get all logs in state
func (self *StateDB) Logs() vm.Logs {
	var logs vm.Logs
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}
// add refund to state temporarily
// Deprecated
func (self *StateDB) AddRefund(gas *big.Int) {
	self.journal.JournalList = append(self.journal.JournalList, &RefundChange{Prev: new(big.Int).Set(self.refund)})
	self.refund.Add(self.refund, gas)
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existant
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.GetStateObject(addr)
	return so == nil || so.empty()
}
// Get account by address
func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	return self.GetStateObject(addr)
}
// Get all account in database
func (self *StateDB) GetAccounts() map[string]vm.Account {
	ret := make(map[string]vm.Account)
	leveldb, ok := self.db.(*hyperdb.LDBDatabase)
	if ok == false {
		return ret
	}

	iter := leveldb.NewIteratorWithPrefix([]byte(accountIdentifier))
	for iter.Next() {
		addr, ok := SplitCompositeAccountKey(iter.Key())
		if ok == false {
			continue
		}
		address := common.BytesToAddress(addr)
		var account Account
		json.Unmarshal(iter.Value(), &account)
		newobj := newObject(self, address, account, self.MarkStateObjectDirty, false, nil)
		ret[address.Hex()] = newobj
	}
	return ret
}
// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}
// Nonce is the contract number state object has deployed
func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}
// code is the contract's code
func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		code := stateObject.Code(self.db)
		key := common.BytesToHash(stateObject.CodeHash())
		self.codeSizeCache.Add(key, len(code))
		return code
	}
	return nil
}
// Get code len
func (self *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return 0
	}
	key := common.BytesToHash(stateObject.CodeHash())
	if cached, ok := self.codeSizeCache.Get(key); ok {
		return cached.(int)
	}
	size := len(stateObject.Code(self.db))
	if stateObject.dbErr == nil {
		self.codeSizeCache.Add(key, size)
	}
	return size
}
// Get code hash
func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}
// get a storage entry in stateObject
func (self *StateDB) GetState(a common.Address, b common.Hash) common.Hash {
	// Prefer `live` object
	var liveObj *StateObject
	if obj := self.stateObjects[a]; obj != nil {
		liveObj = obj
		if obj.deleted {
			return common.Hash{}
		} else {
			value := obj.GetState(b)
			// if storage entry exist in live object's storage cache
			if (value != common.Hash{}) {
				log.Criticalf("get state for %x in live objects, key %x, value %x", a.Hex(), b.Hex(), value.Hex())
				return value
			}
		}
	}
	// find in content cache
	if self.curSeqNo > 0 {
		for i := self.curSeqNo - 1; i >= atomic.LoadUint64(&self.oldestSeqNo); i -= 1 {
			res, existed := self.contentCache.Get(i)
			if existed == false {
				log.Error("load content from content cache failed")
				continue
			}
			content := res.(map[common.Address]*StateObject)
			if obj := content[a]; obj != nil {
				if obj.deleted {
					return common.Hash{}
				} else {
					value := obj.GetState(b)
					if (value != common.Hash{}) {
						log.Criticalf("get state for %x in content cache, key %x, value %x", a.Hex(), b.Hex(), value.Hex())
						if liveObj == nil {
							// save obj itself to current cache
							self.setStateObject(obj)
						} else {
							// save into live obj's cache storage avoid disk cost for next fetch
							liveObj.cachedStorage[b] = value
						}
						return value
					}
				}
			}
		}
	}
	// load from database
	value := GetStateFromDB(self.db, a, b)
	if (value != common.Hash{}) {
		if liveObj == nil {
			log.Criticalf("get state for %x in database, key %x, value %x, add to live state object's storage cache", a.Hex(), b.Hex(), value.Hex())
			// Load the object from the database.
			data, err := self.db.Get(CompositeAccountKey(a.Bytes()))
			if err != nil {
				return common.Hash{}
			}
			var account Account
			err = Unmarshal(data, &account)
			if err != nil {
				return common.Hash{}
			}
			// Insert into the live set.
			obj := newObject(self, a, account, self.MarkStateObjectDirty, true, SetupBucketConfig(self.bktConf.StorageSize, self.bktConf.StorageLevelGroup))
			obj.cachedStorage[b] = value
			self.setStateObject(obj)
		} else {
			// save into live obj's cache storage avoid disk cost for next fetch
			log.Criticalf("get state for %x in database, key %x, value %x, add %x to live objects", a.Hex(), b.Hex(), value.Hex(), a.Hex())
			liveObj.cachedStorage[b] = value
		}
		return value
	}
	log.Criticalf("find state for %x %x failed", a.Hex(), b.Hex())
	return common.Hash{}
}
// check whether an account has been suicide
func (self *StateDB) IsDeleted(addr common.Address) bool {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */
// add balance to an account
func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}
// set balance
func (self *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}
// set nonce
func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}
// set code
func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(kec256Hash.Hash(stateObject.code), code)
	}
}
// set a storage entry to a state object
func (self *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	log.Error("hyper statedb set state start")
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		log.Error("hyper statedb set state find state object in live objects")
		stateObject.SetState(self.db, key, value)
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

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(batch hyperdb.Batch,stateObject *StateObject) []byte {
	addr := stateObject.Address()
	data, err := stateObject.MarshalJSON()
	if err != nil {
		log.Error("marshal stateobject failed", addr.Hex())
	}
	batch.Put(CompositeAccountKey(addr.Bytes()), data)
	return data
}

// deleteStateObject removes the given object from the database
func (self *StateDB) deleteStateObject(batch hyperdb.Batch,stateObject *StateObject) {
	log.Criticalf("delete state object %s during state commit, seqNo #%d", stateObject.address.Hex(), self.curSeqNo)
	stateObject.deleted = true
	addr := stateObject.Address()
	batch.Delete(CompositeAccountKey(addr.Bytes()))
}

// Retrieve a state object given my the address. Returns nil if not found.
func (self *StateDB) GetStateObject(addr common.Address) *StateObject {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			log.Criticalf("search state object %x in the live objects, but it has been suicide", addr)
			return nil
		}
		log.Criticalf("search state object %x in the live objects", addr)
		return obj
	}
	// Load from the content cache
	if self.curSeqNo != 0 {
		for i := self.curSeqNo - 1; i >= atomic.LoadUint64(&self.oldestSeqNo); i -= 1 {
			res, existed := self.contentCache.Get(i)
			if existed == false {
				log.Error("load content from content cache failed")
				continue
			}
			content := res.(map[common.Address]*StateObject)
			if obj := content[addr]; obj != nil {
				if obj.deleted {
					log.Criticalf("search state object %x in the content cache, but it has been suicide", addr)
					return nil
				}
				log.Criticalf("search state object %x in the content cache, add it to live objects", addr)
				self.setStateObject(obj)
				return obj
			}
		}
	}
	// Load the object from the database.
	data, err := self.db.Get(CompositeAccountKey(addr.Bytes()))
	if err != nil {
		log.Critical("no state object been find")
		return nil
	}
	var account Account
	err = Unmarshal(data, &account)
	if err != nil {
		return nil
	}
	// Insert into the live set.
	log.Criticalf("find state object %x in database, add it to live objects", addr)
	obj := newObject(self, addr, account, self.MarkStateObjectDirty, true, SetupBucketConfig(self.bktConf.StorageSize, self.bktConf.StorageLevelGroup))
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *StateObject) {
	self.stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil
func (self *StateDB) GetOrNewStateObject(addr common.Address) *StateObject {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	return stateObject
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
	newobj = newObject(self, addr, Account{}, self.MarkStateObjectDirty, true, SetupBucketConfig(self.bktConf.StorageSize, self.bktConf.StorageLevelGroup))
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		log.Infof("(+) %x\n", addr)
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

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (self *StateDB) Copy() *StateDB {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                self.db,
		root:              self.root,
		codeSizeCache:     self.codeSizeCache,
		stateObjects:      make(map[common.Address]*StateObject, len(self.stateObjectsDirty)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.stateObjectsDirty)),
		refund:            new(big.Int).Set(self.refund),
		logs:              make(map[common.Hash]vm.Logs, len(self.logs)),
		logSize:           self.logSize,
	}
	// Copy the dirty states and logs
	for addr := range self.stateObjectsDirty {
		state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state, state.MarkStateObjectDirty)
		state.stateObjectsDirty[addr] = struct{}{}
	}
	for hash, logs := range self.logs {
		state.logs[hash] = make(vm.Logs, len(logs))
		copy(state.logs[hash], logs)
	}
	return state
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
		log.Error("Revert to snapshot", copy, "failed")
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
		self.journal.JournalList[i].Undo(self, false)
		log.Info("undo operation: %s", self.journal.JournalList[i])
	}
	self.journal.JournalList = self.journal.JournalList[:snapshot]

	// Remove invalidated snapshots from the stack.
	self.validRevisions = self.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
// The return value must not be modified by the caller and will become
// invalid at the next call to AddRefund.
func (self *StateDB) GetRefund() *big.Int {
	return self.refund
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
// Deprecated
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return common.Hash{}
}

// DeleteSuicides flags the suicided objects for deletion so that it
// won't be referenced again when called / queried up on.
//
// DeleteSuicides should not be used for consensus related updates
// under any circumstances.
// Deprecated
func (s *StateDB) DeleteSuicides() {
	// Reset refund so that any used-gas calculations can use this method.
	// s.clearJournalAndRefund()

	for addr := range s.stateObjectsDirty {
		stateObject := s.stateObjects[addr]

		// If the object has been removed by a suicide
		// flag the object as deleted.
		if stateObject.suicided {
			stateObject.deleted = true
		}
		delete(s.stateObjectsDirty, addr)
	}
}

// Commit commits all state changes to the database.
func (s *StateDB) Commit() (common.Hash, error) {
	root, _ := s.CommitBatch(deleteEmptyObjects)
	// IMPORTANT doesn't flush batch util recv commit event for atomic assurance
	return root, nil
}

// CommitBatch commits all state changes to a write batch but does not
// execute the batch. It is used to validate state changes against
// the root hash stored in a block.
func (s *StateDB) CommitBatch(deleteEmptyObjects bool) (root common.Hash, batch hyperdb.Batch) {
	curSeqNo := atomic.LoadUint64(&s.curSeqNo)
	batch = s.FetchBatch(curSeqNo)
	root, _ = s.commit(batch, deleteEmptyObjects)
	return root, batch
}

func (s *StateDB) clearJournalAndRefund(batch hyperdb.Batch) {
	// 1. persist current journal to disk
	curSeqNo := atomic.LoadUint64(&s.curSeqNo)
	if len(s.journal.JournalList) != 0 {
		if curSeqNo == 0 {
			log.Warningf("make sure the seqNo of journal is [%d]", curSeqNo)
		}
		journal, err := s.journal.Marshal()
		if err != nil {
			log.Errorf("marshal seqNo [%d] journal failed", curSeqNo)
		}
		batch.Put(CompositeJournalKey(curSeqNo), journal)
	}
	// 2. clear journal
	s.journal = Journal{}
	s.validRevisions = s.validRevisions[:0]
	// 3. clear refund
	s.refund = new(big.Int)
}

func (s *StateDB) commit(dbw hyperdb.Batch, deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund(dbw)
	var set ChangeSet
	workingSet := bucket.NewKVMap()
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			log.Errorf("seqNo #%d, state object %s been suicide or clearing out for empty", s.curSeqNo, stateObject.address.Hex())
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
			log.Errorf("seqNo #%d, state object %s been updated", s.curSeqNo, stateObject.address.Hex())
			if stateObject.code != nil && stateObject.dirtyCode {
				if err := dbw.Put(CompositeCodeHash(stateObject.Address().Bytes(), stateObject.CodeHash()), stateObject.code); err != nil {
					return common.Hash{}, err
				}
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.Flush(dbw); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			d := s.updateStateObject(dbw, stateObject)
			// Add to change set
			if enableFakeHashFn {
				c := append(stateObject.address.Bytes(), d...)
				set = append(set, c)
			} else {
				workingSet[stateObject.address.Hex()] = d
			}
		}
		delete(s.stateObjectsDirty, addr)
	}
	// Write trie changes.
	// Calculate Hash
	if enableFakeHashFn {
		s.root = SimpleHashFn(s.root, set)
	} else {
		// Use bucket tree instead
		log.Errorf("begin to calculate state db root hash for #%d", s.curSeqNo)
		s.bucketTree.PrepareWorkingSet(workingSet)
		hash, err := s.bucketTree.ComputeCryptoHash()
		if err != nil {
			log.Error("calculate hash for statedb failed")
			return common.Hash{}, err
		}
		s.root = common.BytesToHash(hash)
		s.bucketTree.AddChangesForPersistence(dbw)
	}
	return s.root, err
}

// check bucket root validation
func validateRoot(root common.Hash, curRoot common.Hash) bool {
	if root != curRoot {
		return false
	} else {
		return true
	}
}
