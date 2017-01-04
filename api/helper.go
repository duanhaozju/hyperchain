package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/tree/bucket"
	"hyperchain/common"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"hyperchain/core/state"
	"hyperchain/core"
	"errors"
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