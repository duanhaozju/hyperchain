package db_utils

import "hyperchain/hyperdb/db"

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
