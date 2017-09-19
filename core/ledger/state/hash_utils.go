package state

import (
	"bytes"
	"hyperchain/common"
	"sort"
)

type ChangeSet [][]byte

func (set ChangeSet) Len() int {
	return len(set)
}
func (set ChangeSet) Swap(i, j int) {
	set[i], set[j] = set[j], set[i]
}
func (set ChangeSet) Less(i, j int) bool {
	return bytes.Compare(set[i], set[j]) == -1
}
func SimpleHashFn(root common.Hash, set ChangeSet) common.Hash {
	sort.Sort(set)
	return kec256Hash.Hash([]interface{}{
		root,
		set,
	})
}
