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
)


const (
	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000

	// whether turn on fake hash function
	enableFakeHashFn =  true
	// whether to remove empty stateObject
	deleteEmptyObjects = false
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
	journal        journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex
}

// Create a new state from a given root
func New(root common.Hash, db hyperdb.Database) (*StateDB, error) {
	csc, _ := lru.New(codeSizeCacheSize)
	return &StateDB{
		db:                db,
		root:              root,
		codeSizeCache:     csc,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]vm.Logs),
	}, nil
}

// New creates a new statedb by reusing journalled data to avoid costly
// disk io.
func (self *StateDB) New(root common.Hash) (*StateDB, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	return &StateDB{
		db:                self.db,
		codeSizeCache:     self.codeSizeCache,
		root:              root,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]vm.Logs),
	}, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.root = root
	self.stateObjects = make(map[common.Address]*StateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash]vm.Logs)
	self.logSize = 0
	self.clearJournalAndRefund()
	return nil
}


func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (self *StateDB) AddLog(log *vm.Log) {
	self.journal = append(self.journal, &addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) vm.Logs {
	return self.logs[hash]
}

func (self *StateDB) Logs() vm.Logs {
	var logs vm.Logs
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

func (self *StateDB) AddRefund(gas *big.Int) {
	self.journal = append(self.journal, &refundChange{prev: new(big.Int).Set(self.refund)})
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

func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	return self.GetStateObject(addr)
}

func (self *StateDB) GetAccounts() map[string]vm.Account {
	ret := make(map[string]vm.Account)
	// TODO be more elegant
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
		newobj := newObject(self, address, account, self.MarkStateObjectDirty)
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

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

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

func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (self *StateDB) GetState(a common.Address, b common.Hash) common.Hash {
	stateObject := self.GetStateObject(a)
	if stateObject != nil {
		return stateObject.GetState(self.db, b)
	}
	return common.Hash{}
}

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

func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (self *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(kec256Hash.Hash(stateObject.code), code)
	}
}

func (self *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
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
	self.journal = append(self.journal, &suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: new(big.Int).Set(stateObject.Balance()),
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)
	return true
}

//
// Setting, updating & deleting state object methods
//

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(stateObject *StateObject) {
	addr := stateObject.Address()
	data, err := stateObject.MarshalJSON()
	if err != nil {
		log.Error("marshal stateobject failed", addr.Hex())
	}
	self.db.Put(CompositeAccountKey(addr.Bytes()), data)
}

// deleteStateObject removes the given object from the database
func (self *StateDB) deleteStateObject(stateObject *StateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.db.Delete(addr.Bytes())
}

// Retrieve a state object given my the address. Returns nil if not found.
func (self *StateDB) GetStateObject(addr common.Address) (stateObject *StateObject) {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	data, err := self.db.Get(CompositeAccountKey(addr.Bytes()))
	if err != nil {
		return nil
	}
	var account Account
	err = UnmarshalJSON(data, &account)
	if err != nil {
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, account, self.MarkStateObjectDirty)
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
	newobj = newObject(self, addr, Account{}, self.MarkStateObjectDirty)
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		log.Infof("(+) %x\n", addr)
		self.journal = append(self.journal, &createObjectChange{account: &addr})
	} else {
		self.journal = append(self.journal, &resetObjectChange{prev: prev})
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
	self.validRevisions = append(self.validRevisions, revision{id, len(self.journal)})
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
	for i := len(self.journal) - 1; i >= snapshot; i-- {
		self.journal[i].undo(self)
		log.Error("UNDO", self.journal[i])
	}
	self.journal = self.journal[:snapshot]

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
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return common.Hash{}
}

// DeleteSuicides flags the suicided objects for deletion so that it
// won't be referenced again when called / queried up on.
//
// DeleteSuicides should not be used for consensus related updates
// under any circumstances.
func (s *StateDB) DeleteSuicides() {
	// Reset refund so that any used-gas calculations can use this method.
	s.clearJournalAndRefund()

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
func (s *StateDB) Commit() (root common.Hash, err error) {
	root, batch := s.CommitBatch(deleteEmptyObjects)
	return root, batch.Write()
}

// CommitBatch commits all state changes to a write batch but does not
// execute the batch. It is used to validate state changes against
// the root hash stored in a block.
func (s *StateDB) CommitBatch(deleteEmptyObjects bool) (root common.Hash, batch hyperdb.Batch) {
	batch = s.db.NewBatch()
	root, _ = s.commit(batch, deleteEmptyObjects)

	return root, batch
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = nil
	s.validRevisions = s.validRevisions[:0]
	s.refund = new(big.Int)
}

func (s *StateDB) commit(dbw hyperdb.Batch, deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund()
	var set ChangeSet
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(stateObject)
			d := stateObject.address.Bytes()
			set = append(set, d)
		case isDirty:
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				if err := dbw.Put(CompositeCodeHash(stateObject.Address().Bytes(), stateObject.CodeHash()), stateObject.code); err != nil {
					return common.Hash{}, err
				}
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			batch := s.db.NewBatch()
			if err := stateObject.Flush(batch); err != nil {
				return common.Hash{}, err
			}
			batch.Write()
			// Update the object in the main account trie.
			s.updateStateObject(stateObject)
			// Add to change set
			sd, _ := stateObject.MarshalJSON()
			d := append(stateObject.address.Bytes(), sd...)
			set = append(set, d)
		}
		delete(s.stateObjectsDirty, addr)
	}
	// Write trie changes.
	// Calculate Hash
	if enableFakeHashFn {
		s.root = SimpleHashFn(s.root, set)
	} else {
		// Use bucket tree instead
	}
	return s.root, err
}


