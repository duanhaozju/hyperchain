package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(path string) (Database, error) {
	option := &opt.Options{
		ReadOnly: true,
	}
	db, err := leveldb.OpenFile(path, option)
	if err != nil {
		return nil, err
	}
	levelDB := &LevelDB{
		db: db,
	}
	return levelDB, nil
}

func (levelDB *LevelDB) Get(key []byte) ([]byte, error) {
	return levelDB.db.Get(key, nil)
}

func (levelDB *LevelDB) Close() error {
	return levelDB.db.Close()
}

func (levelDB *LevelDB) NewIterator(prefix []byte) Iterator {
	return levelDB.db.NewIterator(util.BytesPrefix(prefix), nil)
}
