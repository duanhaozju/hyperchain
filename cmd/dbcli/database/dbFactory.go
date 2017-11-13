package database

import (
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
)

func DBFactory(db, path string) (Database, error) {
	switch db {
	case "leveldb":
		return NewLevelDB(path)
	default:
		return nil, constant.ErrInvalidDBParams
	}

}
