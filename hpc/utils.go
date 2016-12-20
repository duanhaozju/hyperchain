package hpc

import (
	"hyperchain/core/vm"
	"hyperchain/core/hyperstate"
	"hyperchain/core/state"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"errors"
	"hyperchain/tree/bucket"
)

func GetStateInstance(root common.Hash, db hyperdb.Database, stateType string, bucketConf bucket.Conf) (vm.Database, error) {
	switch stateType {
	case "rawstate":
		return state.New(root, db)
	case "hyperstate":
		return hyperstate.New(root, db, bucketConf)
	default:
		return nil, errors.New("no state type specified")
	}
}

