package hpc

import (
	"errors"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/hyperstate"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/tree/bucket"
)

func GetStateInstance(root common.Hash, db hyperdb.Database, stateType string, bucketConf bucket.Conf) (vm.Database, error) {
	switch stateType {
	case "rawstate":
		return state.New(root, db)
	case "hyperstate":
		height := core.GetHeightOfChain()
		return hyperstate.New(root, db, bucketConf, height)
	default:
		return nil, errors.New("no state type specified")
	}
}
