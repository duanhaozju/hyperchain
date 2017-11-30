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
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/crypto"
	//"github.com/hyperchain/hyperchain/hyperdb/mdb"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
)

const (
	modulePrefix        = "kvlog."
	entryPrefix         = "entry."
	lastSetPrefix       = "lastSet"
	lastBlockNumPrefix  = "lastBlockNum"
	lastBlockHashPrefix  = "lastBlockHash"
	checkpointMapPrefix = "checkpointMap"
	genesis             = "genesis"
)

// kvLoggerImpl implements the OpLog interface
type kvLoggerImpl struct {
	lastSet          uint64 // The last set entry's index
	lastBlockNum     uint64 // The last set entry's index
	lastBlockHash    string
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
	db, err := hyperdb.GetOrCreateDatabase(config, kvLogger.namespace, hcom.DBNAME_OPLOG)
	//db, err := mdb.NewMemDatabase(kvLogger.namespace)
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
func (kv *kvLoggerImpl) Append(entry *oplog.LogEntry) (uint64, error) {

	kv.mu.Lock()
	defer kv.mu.Unlock()

	var err error
	rollback := &event.RollbackEvent{}
	entry.Lid = kv.lastSet + 1
	if entry.Type == oplog.LogEntry_RollBack {
		if err := proto.Unmarshal(entry.Payload, rollback); err != nil {
			kv.logger.Errorf("Append, unmarshal error: can not unmarshal event.RollbackEvent", err)
			return 0, ErrUnmarshal
		}
		rollback.Lid = entry.Lid
		entry.Payload, err = proto.Marshal(rollback)
		if err != nil {
			kv.logger.Errorf("Append, marshal error: can not marshal event.RollbackEvent", err)
			return 0, ErrMarshal
		}
	}

	raw, err := proto.Marshal(entry)
	if err != nil {
		kv.logger.Errorf("Append, marshal error: can not marshal oplog.LogEntry", err)
		return 0, ErrMarshal
	}
	key := fmt.Sprintf("%s%s%020d", modulePrefix, entryPrefix, entry.Lid)
	if err = kv.db.Put([]byte(key), raw); err == nil {
		kv.lastSet++
		kv.storeLastSet()
		kv.cache[entry.Lid] = entry

		if entry.Type == oplog.LogEntry_TransactionList {
			kv.lastBlockNum++
			kv.lastBlockHash = crypto.Keccak256Hash(entry.Payload).Hex()
			kv.storeLastBlockNum()
			kv.storeLastBlockHash()
			if kv.lastBlockNum % kv.checkpointPeriod == 0 {
				kv.checkpointMap[kv.lastBlockNum] = kv.lastSet
				kv.storeCheckpointMap()
			}
		} else if entry.Type == oplog.LogEntry_RollBack {
			kv.lastBlockNum = rollback.SeqNo
			_, entry2, err := kv.getBySeqNo(rollback.SeqNo)
			if err != nil {
				kv.logger.Errorf("cannot rollback to %d", rollback.SeqNo)
				return 0, ErrRollback
			}
			kv.lastBlockHash = crypto.Keccak256Hash(entry2.Payload).Hex()
			kv.storeLastBlockNum()
			kv.storeLastBlockHash()
			h := kv.lastBlockNum / kv.checkpointPeriod * kv.checkpointPeriod
			for i := range kv.checkpointMap {
				if i > h {
					delete(kv.checkpointMap, i)
				}
			}
			kv.storeCheckpointMap()
		}
		return entry.Lid, nil
	}
	kv.logger.Errorf("Cannot append entry in opLog", err)
	return 0, ErrAppendFail
}

// Fetch get an entry by lid. If this entry is in memory, it can be read in cache.
func (kv *kvLoggerImpl) Fetch(lid uint64) (*oplog.LogEntry, error) {

	kv.mu.Lock()
	defer kv.mu.Unlock()
	if lid > kv.lastSet {
		kv.logger.Debugf("lid %d is too large", lid)
		return nil, ErrLidTooLarge
	}

	if entry, ok := kv.cache[lid]; ok {
		return entry, nil
	}
	key := fmt.Sprintf("%s%s%020d", modulePrefix, entryPrefix, lid)
	raw, err := kv.db.Get([]byte(key))
	if err != nil {
		kv.logger.Errorf("Cannot fetch entry from opLog, lid : %d", lid)
		return nil, ErrNoLid
	}

	entry := &oplog.LogEntry{}
	if err = proto.Unmarshal(raw, entry); err != nil {
		kv.logger.Errorf("Fetch, unmarshal error: can not unmarshal oplog.LogEntry", err)
		return nil, ErrUnmarshal
	} else {
		return entry, nil
	}
}

// Reset set lastSet to a previous number, and later entry would be appended from here.
func (kv *kvLoggerImpl) Reset(seqNo uint64) error {

	//kv.mu.Lock()
	//defer kv.mu.Unlock()
	//
	//if kv.lastBlockNum < seqNo {
	//	kv.logger.Errorf("This seqNo is to large")
	//	return ErrSeqNoTooLarge
	//}
	//lid, _, err := kv.getBySeqNo(seqNo)
	//if err != nil {
	//	return err
	//}
	//kv.lastBlockNum = seqNo
	//kv.lastSet = lid
	//
	//kv.storeLastSet()
	//kv.storeLastBlockNum()
	return nil
}

func (kv *kvLoggerImpl) getBySeqNo(seqNo uint64) (uint64, *oplog.LogEntry, error) {

	if seqNo == 0 || seqNo > kv.lastBlockNum {
		return 0, nil, nil
	}

	checkpoint := uint64((int(seqNo/kv.checkpointPeriod) + 1) * int(kv.checkpointPeriod))
	var earlistLid uint64
	var earlistSeqNo uint64
	if checkpoint < kv.lastBlockNum {
		lid, ok := kv.checkpointMap[checkpoint]
		if !ok {
			kv.logger.Errorf("Not contain this checkpoint: %d in opLog", checkpoint)
		}
		earlistLid = lid
		earlistSeqNo = checkpoint
	} else {
		earlistLid = kv.lastSet
		earlistSeqNo = kv.lastBlockNum
	}
	it := kv.Iterator()

	if !it.Seek(earlistLid) {
		kv.logger.Errorf("Cannot find lid with %d in database", earlistLid)
		return 0, nil, ErrNoLid
	}
	for true {
		if earlistSeqNo < seqNo {
			break
		}
		entry := &oplog.LogEntry{}
		if err := proto.Unmarshal(it.Value(), entry); err != nil {
			kv.logger.Errorf("find, unmarshal error: can not unmarshal oplog.LogEntry", err)
			return 0, nil, ErrUnmarshal
		} else {
			if entry.Type == oplog.LogEntry_TransactionList {
				e := &event.TransactionBlock{}
				if err := proto.Unmarshal(entry.Payload, e); err != nil {
					kv.logger.Errorf("find, unmarshal error: can not unmarshal ValidationEvent", err)
					return 0, nil, ErrUnmarshal
				}
				if earlistSeqNo != e.SeqNo {
					kv.logger.Errorf("find, seqNo didn't match")
					return 0, nil, ErrMismatch
				}
				if e.SeqNo == seqNo {
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

func (kv *kvLoggerImpl) GetLidBySeqNo(seqNo uint64) (uint64, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	lid, _, err := kv.getBySeqNo(seqNo)
	if err != nil {
		return 0, err
	}
	return lid, nil
}

func (kv *kvLoggerImpl) GetLastBlockNum() uint64 {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	return kv.lastBlockNum
}

func (kv *kvLoggerImpl) GetLastSet() uint64 {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	return kv.lastSet
}

func (kv *kvLoggerImpl) SetStableCheckpoint(id uint64) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if lid, ok := kv.checkpointMap[id]; ok {
		for i := range kv.cache {
			if i < lid {
				delete(kv.cache, i)
			}
		}
	} else {
		kv.logger.Errorf("Cannot find this checkpoint %d in SetStableCheckpoint", id)
		return
	}
	for i := range kv.checkpointMap {
		if i <= id {
			delete(kv.checkpointMap, i)
		}
	}
}

func (kv *kvLoggerImpl) GetHeightAndDigest() (uint64, string, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	return kv.lastBlockNum, kv.lastBlockHash, nil
}

func (kv *kvLoggerImpl) storeLastSet() error {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, kv.lastSet)
	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	if err := kv.db.Put([]byte(key), b); err != nil {
		kv.logger.Errorf("Cannot store lastSet in database.")
		return err
	} else {
		return nil
	}
}

func (kv *kvLoggerImpl) restoreLastSet() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	raw, err := kv.db.Get([]byte(key))
	if err != nil {
		kv.logger.Errorf("Cannot restore lastSet from database.")
		kv.lastSet = uint64(0)
		return err
	}
	kv.lastSet = binary.LittleEndian.Uint64(raw)
	return nil
}

func (kv *kvLoggerImpl) storeLastBlockNum() error {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, kv.lastBlockNum)
	key := fmt.Sprintf("%s%s", modulePrefix, lastBlockNumPrefix)
	if err := kv.db.Put([]byte(key), b); err != nil {
		kv.logger.Errorf("Cannot store lastBlockNum in database.")
		return err
	}
	return nil
}

func (kv *kvLoggerImpl) restoreLastBlockNum() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastBlockNumPrefix)
	raw, err := kv.db.Get([]byte(key))
	if err != nil {
		kv.logger.Errorf("Cannot restore lastBlockNum from database.")
		kv.lastBlockNum = uint64(0)
		return err
	}
	kv.lastBlockNum = binary.LittleEndian.Uint64(raw)
	return nil
}

func (kv *kvLoggerImpl) storeLastBlockHash() error {
	key := fmt.Sprintf("%s%s",modulePrefix, lastBlockHashPrefix)
	if err := kv.db.Put([]byte(key), []byte(kv.lastBlockHash)); err != nil {
		kv.logger.Errorf("Cannot store lastBlockHash in database.")
		return err
	}
	return nil
}

func (kv *kvLoggerImpl) restoreLastBlockHash() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastBlockHashPrefix)
	raw, err := kv.db.Get([]byte(key))
	if err != nil {
		kv.logger.Errorf("Cannot restore lastBlockHash from database.")
		kv.lastBlockHash = ""
		return err
	}
	kv.lastBlockHash = string(raw)
	return nil
}

func (kv *kvLoggerImpl) storeCheckpointMap() error {

	checkpointMap := &oplog.CMap{Map: kv.checkpointMap}
	raw, err := proto.Marshal(checkpointMap)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s", modulePrefix, checkpointMapPrefix)
	if err := kv.db.Put([]byte(key), raw); err != nil {
		kv.logger.Errorf("Cannot store checkpointMap in database.")
		return err
	} else {
		return nil
	}
}

func (kv *kvLoggerImpl) restoreCheckpointMap() error {

	key := fmt.Sprintf("%s%s", modulePrefix, checkpointMapPrefix)
	raw, err := kv.db.Get([]byte(key))
	if err != nil {
		kv.logger.Error("Cannot restore checkpointMap from database.")
		kv.checkpointMap = make(map[uint64]uint64)
		return err
	}
	checkpointMap := &oplog.CMap{}
	err = proto.Unmarshal(raw, checkpointMap)
	if err != nil {
		kv.logger.Errorf("restoreCheckpointMap, unmarshal error: can not unmarshal cMap", err)
		kv.checkpointMap = make(map[uint64]uint64)
		return err
	}
	kv.checkpointMap = checkpointMap.Map
	return nil
}

func (kv *kvLoggerImpl) restoreKvLogger() {
	kv.restoreLastSet()
	kv.restoreLastBlockNum()
	kv.restoreLastBlockHash()
	kv.restoreCheckpointMap()
}

// Iterator implements the Iterator interface, could traverse the logger.
type Iterator struct {
	it db.Iterator
}

func (kv *kvLoggerImpl) Iterator() op.Iterator {
	key := fmt.Sprintf("%s%s", modulePrefix, entryPrefix)
	it := &Iterator{}
	it.it = kv.db.NewIterator([]byte(key))
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
