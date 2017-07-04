package common

import cm "hyperchain/common"

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
	if address.Hex() == cm.BytesToAddress(cm.LeftPadBytes([]byte{1}, 20)).Hex() || address.Hex() == cm.BytesToAddress(cm.LeftPadBytes([]byte{2}, 20)).Hex() ||
		address.Hex() == cm.BytesToAddress(cm.LeftPadBytes([]byte{3}, 20)).Hex() || address.Hex() == cm.BytesToAddress(cm.LeftPadBytes([]byte{4}, 20)).Hex() {
		return true
	} else {
		return false
	}
}
