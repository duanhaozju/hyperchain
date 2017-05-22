package db_utils

import (
	"hyperchain/hyperdb/db"
	"strconv"
)

func DeleteAllJournals(db db.Database, batch db.Batch, flush, sync bool) error {
	iter := db.NewIterator(JournalPrefix)
	defer iter.Release()
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	err := iter.Error()
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return err
}

func PersistSnapshot(batch db.Batch, number uint64, buf []byte, flush, sync bool) error {
	// check pointer value
	if buf == nil || batch == nil {
		return EmptyPointerErr
	}
	keyNum := strconv.FormatUint(number, 10)
	if err := batch.Put(append(SnapshotPrefix, []byte(keyNum)...), buf); err != nil {
		return err
	}
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func GetSnapshot(db db.Database, number uint64) ([]byte, error) {
	keyNum := strconv.FormatUint(number, 10)
	if buf, err := db.Get(append(SnapshotPrefix, []byte(keyNum)...)); err != nil {
		return nil, err
	} else {
		return buf, err
	}
}

