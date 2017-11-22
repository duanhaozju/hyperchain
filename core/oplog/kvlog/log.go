package kvlog

import (
	"fmt"
	"errors"
	"encoding/binary"

	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/core/oplog/proto"
	op "github.com/hyperchain/hyperchain/core/oplog"

	"github.com/golang/protobuf/proto"
	"sync/atomic"
)

const (
	modulePrefix = "kvlog."
	entryPrefix = "entry."
	lastSetPrefix = "lastSet"
)

// kvLoggerImpl implements the OpLog interface
type kvLoggerImpl struct {
	lastSet	uint64 // The last set entry's index
	cache	map[uint64]*oplog.LogEntry // A cache to store some entry in memory
	db		db.Database // A database to store entries
}

// New initiate a kvLoggerImpl
func New(db db.Database) *kvLoggerImpl {
	
	logger := &kvLoggerImpl{
		db:		db,
		cache:	make(map[uint64]*oplog.LogEntry),
	}
	err := logger.restoreLastSet()
	if err != nil {
		atomic.StoreUint64(&logger.lastSet, uint64(0))
	}
	return logger
}

// Append add an entry to logger
func (logger *kvLoggerImpl) Append(entry *oplog.LogEntry) error {

	lastSet := atomic.LoadUint64(&logger.lastSet)
	if entry.Lid != lastSet + 1 {
		return errors.New(fmt.Sprint("not matched with lastSet"))
	}

	raw, err := proto.Marshal(entry)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s%d", modulePrefix, entryPrefix, entry.Lid)
	if err = logger.db.Put([]byte(key), raw); err == nil {
		logger.cache[entry.Lid] = entry

		// TODO How to delete these massage in cache
		if entry.Lid % 10 == 0 {
			for i := entry.Lid - 49; i <= entry.Lid - 40 && i > 0; i++ {
				delete(logger.cache, i)
			}
		}

		atomic.AddUint64(&logger.lastSet, 1)
		logger.storeLastSet()
	}
	return err
}

// Fetch get an entry by lid. If this entry is in memory, it can be read in cache.
func (logger *kvLoggerImpl) Fetch(lid uint64) (*oplog.LogEntry, error) {

	lastSet := atomic.LoadUint64(&logger.lastSet)
	if lid > lastSet {
		return nil, errors.New(fmt.Sprint("lid is too large"))
	}

	if entry, ok := logger.cache[lid]; ok {
		return entry, nil
	}
	key := fmt.Sprintf("%s%s%d", modulePrefix, entryPrefix, lid)
	raw, err := logger.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	entry := &oplog.LogEntry{}
	if err = proto.Unmarshal(raw, entry); err != nil{
		return nil, err
	} else {
		return entry, nil
	}
}

// Reset set lastSet to a previous number, and later entry would be appended from here.
func (logger *kvLoggerImpl) Reset(lid uint64) error {

	lastSet := atomic.LoadUint64(&logger.lastSet)
	if lid > lastSet {
		return errors.New(fmt.Sprint("lid is too large"))
	}

	atomic.StoreUint64(&logger.lastSet, lid)
	logger.storeLastSet()
	return nil
}

//
func (logger *kvLoggerImpl) storeLastSet() error {

	b := make([]byte, 8)
	lastSet := atomic.LoadUint64(&logger.lastSet)
	binary.LittleEndian.PutUint64(b, lastSet)
	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	if err := logger.db.Put([]byte(key), b); err != nil {
		return err
	} else {
		return nil
	}
}

func (logger *kvLoggerImpl) restoreLastSet() error {

	key := fmt.Sprintf("%s%s", modulePrefix, lastSetPrefix)
	raw, err := logger.db.Get([]byte(key))
	if err != nil {
		return err
	}
	atomic.StoreUint64(&logger.lastSet, binary.LittleEndian.Uint64(raw))
	return nil
}

// Iterator implements the Iterator interface, could traverse the logger.
type Iterator struct {
	it db.Iterator
}

func (logger *kvLoggerImpl) Iterator(prefix []byte) op.Iterator {
	it := &Iterator{}
	it.it = logger.db.NewIterator(prefix)
	return it
}

func (it *Iterator) Key() []byte {
	return it.it.Key()
}

func (it *Iterator) Value() []byte {
	return it.it.Value()
}

func (it *Iterator) Seek(key []byte) bool {
	return it.it.Seek(key)
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

//func (logger *kvLoggerImpl) FetchEntries(lowLid, highLid uint64) (map[uint64]*oplog.LogEntry, error) {
//
//	entries := make(map[uint64]*oplog.LogEntry)
//	prefixRaw := []byte(modulePrefix + entryPrefix)
//	it := logger.db.NewIterator(prefixRaw)
//	if it == nil {
//		err := errors.New(fmt.Sprint("Can't get Iterator"))
//		return nil, err
//	}
//	if !it.Seek(prefixRaw) {
//		err := errors.New(fmt.Sprintf("Cannot find key with %s in database", prefixRaw))
//		return nil, err
//	}
//	for bytes.HasPrefix(it.Key(), prefixRaw) {
//		key := string(it.Key())
//		keyVal, err := strconv.ParseUint(key[len(modulePrefix) + len(entryPrefix):], 10, 64)
//		if err != nil {
//			return nil, err
//		}
//		if keyVal >= lowLid && keyVal <= highLid {
//			entry := &oplog.LogEntry{}
//			if err = proto.Unmarshal(it.Value(), entry); err != nil{
//				return nil, err
//			} else {
//				entries[keyVal] =  entry
//			}
//		}
//		if !it.Next() {
//			break
//		}
//	}
//	it.Release()
//	if len(entries) != int(highLid - lowLid + 1) {
//		return nil, errors.New(fmt.Sprint("entries' length don't match"))
//	}
//	return entries, nil
//}