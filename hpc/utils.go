package hpc

import (
	"hyperchain/core/vm"
	"hyperchain/core/hyperstate"
	"hyperchain/core/state"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"errors"
)

func GetStateInstance(root common.Hash, db hyperdb.Database, stateType string) (vm.Database, error) {
	switch stateType {
	case "rawstate":
		return state.New(root, db)
	case "hyperstate":
		return hyperstate.New(root, db)
	default:
		return nil, errors.New("no state type specified")
	}
}

