package common

import cm "hyperchain/common"

var precompiledAccount = map[string]struct{}{
	cm.BytesToAddress(cm.LeftPadBytes([]byte{1}, 20)).Hex(): struct{}{},  // ECRECOVER
	cm.BytesToAddress(cm.LeftPadBytes([]byte{2}, 20)).Hex(): struct{}{},  // SHA256
	cm.BytesToAddress(cm.LeftPadBytes([]byte{3}, 20)).Hex(): struct{}{},  // RIPEMD160
	cm.BytesToAddress(cm.LeftPadBytes([]byte{4}, 20)).Hex(): struct{}{},  // MEMCPY
}


func RetrieveSnapshotFileds() []string {
	return []string{
		// world state related
		"-account",
		"-storage",
		"-code",
		// bucket tree related
		"-bucket",
		"BucketNode",
	}
}

func IsPrecompiledAccount(address cm.Address) bool {
	if _, exist := precompiledAccount[address.Hex()]; exist {
		return true
	}
	return false
}


