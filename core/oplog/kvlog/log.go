package kvlog

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/hyperchain/hyperchain/common"
	op "github.com/hyperchain/hyperchain/core/oplog"
	"github.com/hyperchain/hyperchain/core/oplog/proto"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/manager/event"

	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
	"github.com/op/go-logging"
)

const (
	modulePrefix        = "kvlog."
	entryPrefix         = "entry."
	lastSetPrefix       = "lastSet"
	lastCommitPrefix    = "lastCommit"
	checkpointMapPrefix = "checkpointMap"
)

// kvLoggerImpl implements the OpLog interface
type kvLoggerImpl struct {
	lastSet          uint64 // The last set entry's index
	lastCommit       uint64 // The last set entry's index
	checkpointPeriod uint64
	checkpointMap    map[uint64]uint64

	namespace string
	mu        *sync.Mutex
	cache     map[uint64]*oplog.LogEntry // A cache to store some entry in memory
	db        db.Database                // A database to store entries
	logger    *logging.Logger
}

// New initiate a kvLoggerImpl
func New(config *common.Config) *kvLoggerImpl {

	kvLogger := &kvLoggerImpl{
		namespace:        config.GetString(common.NAMESPACE),
		checkpointPeriod: uint64(config.GetInt64("consensus.rbft.k")),
		mu:               new(sync.Mutex),
		cache:            make(map[uint64]*oplog.LogEntry),
	}
	//db, err := hyperdb.GetOrCreateDatabase(config, kvLogger.namespace, hcom.DBNAME_OPLOG)
	db, err := mdb.NewMemDatabase(kvLogger.namespace)
	if err != nil {
		kvLogger.logger.Errorf("get opLog db by namespace: %s failed.", kvLogger.namespace)
		return nil
	}
	kvLogger.db = db
	kvLogger.logger = common.GetLogger(kvLogger.namespace, "opLog")
	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()
	kvLogger.restoreKvLogger()
	return kvLogger
}

// Append add an entry to logger
func (kvLogger *kvLoggerImpl) Append(entry *oplog.LogEntry) error {

	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()

	entry.Lid = kvLogger.lastSet + 1
	raw, err := proto.Marshal(entry)
	if err != nil {
		kvLogger.logger.Errorf("Append, marshal error: can not marshal oplog.LogEntry", err)
		return ErrMarshal
	}
	key := fmt.Sprintf("%s%s%020d", modulePrefix, entryPrefix, entry.Lid)
	if err = kvLogger.db.Put([]byte(key), raw); err == nil {
		kvLogger.lastSet++
		kvLogger.storeLastSet()
		kvLogger.cache[entry.Lid] = entry

		// TODO How to delete these massage in cache
		if entry.Lid%kvLogger.checkpointPeriod == 0 {
			for i := entry.Lid - 5*kvLogger.checkpointPeriod + 1; i <= entry.Lid-4*kvLogger.checkpointPeriod && i > 0; i++ {
				delete(kvLogger.cache, i)
			}
		}
		if entry.Type == oplog.LogEntry_TransactionList {
			kvLogger.lastCommit++
			kvLogger.storeLastCommit()
			if kvLogger.lastCommit%kvLogger.checkpointPeriod == 0 {
				kvLogger.checkpointMap[kvLogger.lastCommit] = kvLogger.lastSet
				kvLogger.storeCheckpointMap()
			}
		}
		return nil
	}
	kvLogger.logger.Errorf("Cannot append entry in opLog", err)
	return ErrAppendFail
}

// Fetch get an entry by lid. If this entry is in memory, it can be read in cache.
func (kvLogger *kvLoggerImpl) Fetch(lid uint64) (*oplog.LogEntry, error) {

	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()
	if lid > kvLogger.lastSet {
		kvLogger.logger.Debugf("lid %d is too large", lid)
		return nil, ErrLidTooLarge
	}

	if entry, ok := kvLogger.cache[lid]; ok {
		return entry, nil
	}
	key := fmt.Sprintf("%s%s%020d", modulePrefix, entryPrefix, lid)
	raw, err := kvLogger.db.Get([]byte(key))
	if err != nil {
		kvLogger.logger.Errorf("Cannot fetch entry from opLog, lid : %d", lid)
		return nil, ErrNoLid
	}

	entry := &oplog.LogEntry{}
	if err = proto.Unmarshal(raw, entry); err != nil {
		kvLogger.logger.Errorf("Fetch, unmarshal error: can not unmarshal oplog.LogEntry", err)
		return nil, ErrUnmarshal
	} else {
		return entry, nil
	}
}

// Reset set lastSet to a previous number, and later entry would be appended from here.
func (kvLogger *kvLoggerImpl) Reset(seqNo uint64) error {

	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()

	if kvLogger.lastCommit < seqNo {
		kvLogger.logger.Errorf("This seqNo is to large")
		return ErrSeqNoTooLarge
	}
	lid, _, err := kvLogger.getBySeqNo(seqNo)
	if err != nil {
		return err
	}
	kvLogger.lastCommit = seqNo
	kvLogger.lastSet = lid

	kvLogger.storeLastSet()
	kvLogger.storeLastCommit()
	return nil
}

func (kvLogger *kvLoggerImpl) getBySeqNo(seqNo uint64) (uint64, *oplog.LogEntry, error) {

	if seqNo == 0 {
		return 0, nil, nil
	}

	checkpoint := uint64((int(seqNo/kvLogger.checkpointPeriod) + 1) * int(kvLogger.checkpointPeriod))
	var earlistLid uint64
	var earlistSeqNo uint64
	if checkpoint < kvLogger.lastCommit {
		lid, ok := kvLogger.checkpointMap[checkpoint]
		if !ok {
			kvLogger.logger.Errorf("Not contain this checkpoint: %d in opLog", checkpoint)
		}
		earlistLid = lid
		earlistSeqNo = checkpoint
	} else {
		earlistLid = kvLogger.lastSet
		earlistSeqNo = kvLogger.lastCommit
	}
	it := kvLogger.Iterator()

	if !it.Seek(earlistLid) {
		kvLogger.logger.Errorf("Cannot find lid with %d in database", earlistLid)
		return 0, nil, ErrNoLid
	}
	for true {
		if earlistSeqNo < seqNo {
			break
		}
		entry := &oplog.LogEntry{}
		if err := proto.Unmarshal(it.Value(), entry); err != nil {
			kvLogger.logger.Errorf("find, unmarshal error: can not unmarshal oplog.LogEntry", err)
			return 0, nil, ErrUnmarshal
		} else {
			if entry.Type == oplog.LogEntry_TransactionList {
				event := &event.TransactionBlock{}
				if err := proto.Unmarshal(entry.Payload, event); err != nil {
					kvLogger.logger.Errorf("find, unmarshal error: can not unmarshal ValidationEvent", err)
					return 0, nil, ErrUnmarshal
				}
				if earlistSeqNo != event.SeqNo {
					kvLogger.logger.Errorf("find, seqNo didn't match")
					return 0, nil, ErrMismatch
				}
				if event.SeqNo == seqNo {
					return earlistLid, entry, nil
				}
				earlistSeqNo--
			}
			earlistLid--
		}
		if !it.Prev() {
			break
		}
	}
	return 0, nil, ErrNoSeqNo
}

func (kvLogger *kvLoggerImpl) GetLastBlockNum() uint64 {
	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()
	return kvLogger.lastCommit
}

func (kvLogger *kvLoggerImpl) GetLastCommit() uint64 {
	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()

	return kvLogger.lastSet
}

func (kvLogger *kvLoggerImpl) SetStableCheckpoint(id uint64) {
	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()

	for i := range kvLogger.checkpointMap {
		if i <= id {
			delete(kvLogger.checkpointMap, i)
		}
	}
}

func (kvLogger *kvLoggerImpl) GetHeightAndDigest() (uint64, string, error) {
	kvLogger.mu.Lock()
	defer kvLogger.mu.Unlock()

	n, entry, err := kvLogger.getBySeqNo(kvLogger.lastCommit)
	if err != nil {
		return 0, "", err
	} else if n == 0 {
		return 0, "", nil
	}
	event := &event.TransactionBlock{}
	if err := proto.Unmarshal(entry.Payload, event); err != nil {
		kvLogger.logger.Errorf("find, unmarshal error: can not unmarshal ValidationEvent", err)
		return 0, "", ErrUnmarshal
	}
	return kvLogger.lastCommit, event.Digest, nil
}

func (kvLogger *kvLoggerImpl) storeLastSet() error {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, kvLogger.lastSet)
	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	if err := kvLogger.db.Put([]byte(key), b); err != nil {
		kvLogger.logger.Errorf("Cannot store lastSet in database.")
		return err
	} else {
		return nil
	}
}

func (kvLogger *kvLoggerImpl) restoreLastSet() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	raw, err := kvLogger.db.Get([]byte(key))
	if err != nil {
		kvLogger.logger.Errorf("Cannot restore lastSet from database.")
		return err
	}
	kvLogger.lastSet = binary.LittleEndian.Uint64(raw)
	return nil
}

func (kvLogger *kvLoggerImpl) storeLastCommit() error {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, kvLogger.lastCommit)
	key := fmt.Sprintf("%s%s", modulePrefix, lastCommitPrefix)
	if err := kvLogger.db.Put([]byte(key), b); err != nil {
		kvLogger.logger.Errorf("Cannot store lastCommit in database.")
		return err
	} else {
		return nil
	}
}

func (kvLogger *kvLoggerImpl) restoreLastCommit() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastCommitPrefix)
	raw, err := kvLogger.db.Get([]byte(key))
	if err != nil {
		kvLogger.logger.Errorf("Cannot restore lastCommit from database.")
		return err
	}
	kvLogger.lastCommit = binary.LittleEndian.Uint64(raw)
	return nil
}

func (kvLogger *kvLoggerImpl) storeCheckpointMap() error {

	checkpointMap := &oplog.CMap{Map: kvLogger.checkpointMap}
	raw, err := proto.Marshal(checkpointMap)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s", modulePrefix, checkpointMapPrefix)
	if err := kvLogger.db.Put([]byte(key), raw); err != nil {
		kvLogger.logger.Errorf("Cannot store checkpointMap in database.")
		return err
	} else {
		return nil
	}
}

func (kvLogger *kvLoggerImpl) restoreCheckpointMap() error {

	key := fmt.Sprintf("%s%s", modulePrefix, checkpointMapPrefix)
	raw, err := kvLogger.db.Get([]byte(key))
	if err != nil {
		kvLogger.logger.Error("Cannot restore checkpointMap from database.")
		return err
	}
	checkpointMap := &oplog.CMap{}
	err = proto.Unmarshal(raw, checkpointMap)
	if err != nil {
		kvLogger.logger.Errorf("restoreCheckpointMap, unmarshal error: can not unmarshal cMap", err)
		return err
	}
	kvLogger.checkpointMap = checkpointMap.Map
	return nil
}

func (kvLogger *kvLoggerImpl) restoreKvLogger() {
	err := kvLogger.restoreLastSet()
	if err != nil {
		kvLogger.lastSet = uint64(0)
	}
	err = kvLogger.restoreLastCommit()
	if err != nil {
		kvLogger.lastCommit = uint64(0)
	}
	err = kvLogger.restoreCheckpointMap()
	if err != nil {
		kvLogger.checkpointMap = make(map[uint64]uint64)
	}
}

// Iterator implements the Iterator interface, could traverse the logger.
type Iterator struct {
	it db.Iterator
}

func (kvLogger *kvLoggerImpl) Iterator() op.Iterator {
	key := fmt.Sprintf("%s%s", modulePrefix, entryPrefix)
	it := &Iterator{}
	it.it = kvLogger.db.NewIterator([]byte(key))
	return it
}

func (it *Iterator) Key() []byte {
	return it.it.Key()
}

func (it *Iterator) Value() []byte {
	return it.it.Value()
}

func (it *Iterator) Seek(lid uint64) bool {
	key := fmt.Sprintf("%s%s%020d", modulePrefix, entryPrefix, lid)
	return it.it.Seek([]byte(key))
}

func (it *Iterator) Next() bool {
	return it.it.Next()
}

func (it *Iterator) Prev() bool {
	return it.it.Prev()
}

func (it *Iterator) Error() error {
	return it.it.Error()
}

func (it *Iterator) Release() {
	it.it.Release()
}
