// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chain

import (
	"encoding/json"
	com "github.com/hyperchain/hyperchain/core/common"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"strconv"
)

// DeleteAllJournals deletes all the journals in database.
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

// DeleteJournalInRange deletes journals in range [start, end).
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

// PersistSnapshotMeta persists the snapshot meta into database.
func PersistSnapshotMeta(batch db.Batch, meta *com.Manifest, flush, sync bool) error {
	if batch == nil || meta == nil {
		return ErrEmptyPointer
	}
	blob, err := json.Marshal(meta)
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

// GetSnapshotMeta gets the snapshot meta with given namespace.
func GetSnapshotMeta(namespace string) (*com.Manifest, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	return GetSnapshotMetaFunc(db)
}

// GetSnapshotMetaFunc gets the snapshot meta with given db handler.
func GetSnapshotMetaFunc(db db.Database) (*com.Manifest, error) {
	blob, err := db.Get([]byte(SnapshotPrefix))
	if err != nil || len(blob) == 0 {
		return nil, err
	}
	var meta com.Manifest
	err = json.Unmarshal(blob, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}
