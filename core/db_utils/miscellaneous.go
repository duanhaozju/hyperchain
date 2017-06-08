package db_utils

import (
	"hyperchain/hyperdb/db"
	"strconv"
	"hyperchain/common"
	"encoding/json"
	"hyperchain/hyperdb"
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

func PersistSnapshotMeta(batch db.Batch, meta *common.Manifest, flush, sync bool) error  {
	if batch == nil || meta == nil {
		return EmptyPointerErr
	}
	blob,  err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := batch.Put([]byte(SnapshotPrefix), blob); err != nil {
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

func GetSnapshotMeta(namespace string) (*common.Manifest, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetSnapshotMetaFunc(db)
}

func GetSnapshotMetaFunc(db db.Database) (*common.Manifest, error) {
	blob, err := db.Get([]byte(SnapshotPrefix))
	if err != nil || len(blob) == 0 {
		return nil, err
	}
	var meta common.Manifest
	err = json.Unmarshal(blob, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}


