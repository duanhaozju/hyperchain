package executor

import (
	"hyperchain/core/types"
	"hyperchain/core/vm/evm"
)


var RemoveLessThan = func(key interface{}, iterKey interface{}) bool {
	id := key.(uint64)
	iterId := iterKey.(uint64)
	if id >= iterId {
		return true
	}
	return false
}

func RetrieveLogs(r *types.Receipt, vmType int32) (interface{}, error) {
	switch vmType {
	case 0:
		// EVM
		return evm.DecodeLogs((*r).Logs)
	case 1:
		// JVM
		return nil, nil
	default:
		// TODO
		return nil, nil
	}
}

func SetLogs(r *types.Receipt, vmType int32, logs interface{}) error {
	var buf []byte
	var err error
	switch vmType {
	case 0:
		// EVM
		tmp := logs.(evm.Logs)
		buf, err = (&tmp).EncodeLogs()
	case 1:
		// JVM
	}
	if err != nil {
		return err
	}
	r.Logs = buf
	return nil
}
