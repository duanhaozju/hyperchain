//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hleveldb

import (
	"bytes"
	"path"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/hyperchain/hyperchain/common"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/p2p/hts/secimpl"
)

// the Database for LevelDB
// LDBDatabase implements the DataBase interface.
type LDBDatabase struct {
	path      string
	db        *leveldb.DB
	conf      *common.Config
	namespace string
	dbEncrypt bool
}

// NewLDBDataBase creates a new LDBDatabase instance,
// it requires a data storage filepath, a namespace string,
// and it returns *LDBDatabase and an error if it exists.
// the LDBDatabase instance must be closed after using, by calling Close method.
func NewLDBDataBase(conf *common.Config, filepath string, namespace string) (*LDBDatabase, error) {
	opt := hcom.GetLdbConfig(conf)
	db, err := leveldb.OpenFile(filepath, opt)
	return &LDBDatabase{
		path:      filepath,
		db:        db,
		conf:      conf,
		namespace: namespace,
		dbEncrypt: conf.GetBool(hcom.DATA_ENCRYPTION),
	}, err
}

// NewRawLDBDatabase creates a new raw LDBDatabase instance which without any config
// it requires a specified leveldb instance and a namespace literal,
// and it returns *LDBDatabase.
func NewRawLDBDatabase(db *leveldb.DB, namespace string) *LDBDatabase {
	return &LDBDatabase{
		db:        db,
		namespace: namespace,
	}
}

// LDBDatabasePath gets database's storage root directory.
func LDBDatabasePath(conf *common.Config) string {
	return conf.GetString(hcom.LEVEL_DB_ROOT_DIR)
}

// Put inserts a K/V pair to the database,
// it will overwrite the value if the key exists.
func (ldb *LDBDatabase) Put(key []byte, value []byte) error {
	var (
		finalBytes []byte = value
		err        error
	)
	if value != nil && len(value) > 0 {
		if ldb.dbEncrypt {
			var privateKey []byte
			if len(key) > hcom.ENCRYPTED_KEY_LEN {
				privateKey = key[:hcom.ENCRYPTED_KEY_LEN]
			} else {
				privateKey = key
				for i := len(key); i < hcom.ENCRYPTED_KEY_LEN; i++ {
					privateKey = append(privateKey, byte(hcom.ENCRYPTED_KEY_APPENDER))
				}
			}
			finalBytes, err = secimpl.TripleDesEncrypt8(finalBytes, privateKey)
			if err != nil {
				return err
			}
		}
	}

	return ldb.db.Put(key, finalBytes, nil)
}

// Get gets a key's value from the database.
func (ldb *LDBDatabase) Get(key []byte) ([]byte, error) {
	value, err := ldb.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		err = db.DB_NOT_FOUND
	}

	var finalBytes []byte = value
	if value != nil {
		if ldb.dbEncrypt {
			var privateKey []byte

			if len(key) > hcom.ENCRYPTED_KEY_LEN {
				privateKey = key[:hcom.ENCRYPTED_KEY_LEN]
			} else {
				privateKey = key
				for i := len(key); i < hcom.ENCRYPTED_KEY_LEN; i++ {
					privateKey = append(privateKey, byte(hcom.ENCRYPTED_KEY_APPENDER))
				}
			}

			finalBytes, err = secimpl.TripleDesDecrypt8(finalBytes, privateKey)
			if err != nil {
				return nil, err
			}
		}
	}
	return finalBytes, err
}

// Delete deletes the value for the given key.
func (ldb *LDBDatabase) Delete(key []byte) error {
	return ldb.db.Delete(key, nil)
}

// Close closes the LDBDatabase.
func (ldb *LDBDatabase) Close() {
	ldb.db.Close()
}

// MakeSnapshot creates a new snapshot for the database.
func (ldb *LDBDatabase) MakeSnapshot(backupPath string, fields []string) error {
	p := path.Join(common.GetPath(ldb.namespace, LDBDatabasePath(ldb.conf)), backupPath)
	backupDb, err := NewLDBDataBase(ldb.conf, p, ldb.namespace)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	snapshot, err := ldb.db.GetSnapshot()
	if err != nil {
		return err
	}
	defer snapshot.Release()

	for _, field := range fields {
		iter := &Iterator{
			ldb:  ldb,
			iter: snapshot.NewIterator(util.BytesPrefix([]byte(field)), nil),
		}

		for iter.Next() {
			if err := backupDb.Put(iter.Key(), iter.Value()); err != nil {
				iter.Release()
				return err
			}
		}
		iter.Release()
	}
	return nil
}

// Namespace returns database's namespace literal.
func (ldb *LDBDatabase) Namespace() string {
	return ldb.namespace
}

// LDB returns *leveldb.DB instance.
func (ldb *LDBDatabase) LDB() *leveldb.DB {
	return ldb.db
}

// Destroy cleans the whole database.
// Warning: bad performance if too many data in the db.
func (ldb *LDBDatabase) Destroy() error {
	return ldb.DestroyByRange(nil, nil)
}

// DestroyByRange cleans objects in range [begin, end)
func (ldb *LDBDatabase) DestroyByRange(begin, end []byte) error {
	if bytes.Compare(begin, end) > 0 {
		return errors.Errorf("begin key: %v, is larger than end key: %v", begin, end)
	}
	iter := ldb.db.NewIterator(&util.Range{Start: begin, Limit: end}, nil)
	for iter.Next() {
		err := ldb.Delete(iter.Key())
		if err != nil {
			iter.Release()
			return err
		}
	}
	iter.Release()
	return nil
}

// ldbBatch implements the Batch interface.
type ldbBatch struct {
	db  *leveldb.DB
	bat *leveldb.Batch
}

// NewBatch returns a Batch instance.
func (ldb *LDBDatabase) NewBatch() db.Batch {
	return &ldbBatch{
		db:  ldb.db,
		bat: new(leveldb.Batch),
	}
}

// Put appends 'put operation' of the given K/V pair to the batch.
func (b *ldbBatch) Put(key, value []byte) error {
	b.bat.Put(key, value)
	return nil
}

// Delete appends 'delete operation' of the given key to the batch.
func (b *ldbBatch) Delete(key []byte) error {
	b.bat.Delete(key)
	return nil
}

// Write apply the given batch to the DB.
func (b *ldbBatch) Write() error {
	return b.db.Write(b.bat, nil)
}

// Len returns number of records in the batch.
func (b *ldbBatch) Len() int {
	return b.bat.Len()
}

// Iterator implements the Iterator interface.
type Iterator struct {
	ldb  *LDBDatabase
	iter iterator.Iterator
}

// NewIterator returns a Iterator for traversing the database.
func (ldb *LDBDatabase) NewIterator(prefix []byte) db.Iterator {
	iter := ldb.db.NewIterator(util.BytesPrefix(prefix), nil)
	return &Iterator{
		ldb:  ldb,
		iter: iter,
	}
}

// Scan is LDBDatabase's Scan method which scans objects in range [begin, end).
func (ldb *LDBDatabase) Scan(begin, end []byte) db.Iterator {
	iter := ldb.db.NewIterator(&util.Range{Start: begin, Limit: end}, nil)
	return &Iterator{
		ldb:  ldb,
		iter: iter,
	}
}

func (ldb *LDBDatabase) NewIteratorWithPrefix(prefix []byte) db.Iterator {
	iter := ldb.db.NewIterator(util.BytesPrefix(prefix), nil)
	return &Iterator{
		ldb:  ldb,
		iter: iter,
	}
}

func (it *Iterator) Key() []byte { return it.iter.Key() }

func (it *Iterator) Value() []byte {
	var err error

	value := it.iter.Value()
	var finalBytes []byte = value
	if value != nil {
		key := it.iter.Key()
		if it.ldb.dbEncrypt {
			var privateKey []byte

			if len(key) > hcom.ENCRYPTED_KEY_LEN {
				privateKey = key[:hcom.ENCRYPTED_KEY_LEN]
			} else {
				privateKey = key
				for i := len(key); i < hcom.ENCRYPTED_KEY_LEN; i++ {
					privateKey = append(privateKey, byte(hcom.ENCRYPTED_KEY_APPENDER))
				}
			}

			finalBytes, err = secimpl.TripleDesDecrypt8(finalBytes, privateKey)
			if err != nil {
				return nil
			}
		}
	}

	return finalBytes
}

func (i *Iterator) Seek(key []byte) bool { return i.iter.Seek(key) }
func (i *Iterator) Next() bool           { return i.iter.Next() }
func (i *Iterator) Prev() bool           { return i.iter.Prev() }
func (i *Iterator) Error() error         { return i.iter.Error() }
func (i *Iterator) Release()             { i.iter.Release() }
