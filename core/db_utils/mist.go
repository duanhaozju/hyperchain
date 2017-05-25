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

// DeleteJournalInRange delete journals in range [start, end)
func DeleteJournalInRange(batch db.Batch, start uint64, end uint64, flush, sync bool) error {
	for i := start; i < end; i += 1 {
		s := strconv.FormatUint(i, 10)
		key := append([]byte(JournalPrefix), []byte(s)...)
		if err := batch.Delete(key); err != nil {
			return err
		}
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


