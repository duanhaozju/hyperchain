package hyperstate

import (
	"bytes"
)

const storageIdentifier = "-storage"
const accountIdentifier = "-account"
const codeIdentifier = "-code"
const bucketTreeIdentifier = "-bucket"

// ConfigNumBuckets - config name 'numBuckets' as it appears in yaml file
const ConfigNumBuckets = "numBuckets"
// ConfigMaxGroupingAtEachLevel - config name 'maxGroupingAtEachLevel' as it appears in yaml file
const ConfigMaxGroupingAtEachLevel = "maxGroupingAtEachLevel"

/*
	Storage
 */
func CompositeStorageKey(address []byte, key []byte) []byte {
	ret := append([]byte(storageIdentifier), address...)
	ret = append(ret, key...)
	return ret
}

func GetStorageKeyPrefix(address []byte) []byte {
	ret := append([]byte(storageIdentifier), address...)
	return ret
}

func SplitCompositeStorageKey(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(storageIdentifier), address...)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) {
		return key[prefixLen:], true
	} else {
		return nil, false
	}
}
/*
	Code
 */
func CompositeCodeHash(address []byte, codeHash []byte) []byte {
	ret := append([]byte(codeIdentifier), address...)
	return  append(ret, codeHash...)
}
/*
	Account
 */
func CompositeAccountKey(address []byte) []byte {
	return append([]byte(accountIdentifier), address...)
}

func SplitCompositeAccountKey(key []byte) ([]byte, bool) {
	identifierLen := len([]byte(accountIdentifier))
	if bytes.HasPrefix(key, []byte(accountIdentifier)) {
		return key[identifierLen:], true
	} else {
		return nil, false
	}
}
/*
	Bucket Tree
 */
func CompositeStateBucketPrefix() ([]byte, bool) {
	return append([]byte(bucketTreeIdentifier), []byte("-state")...), true
}
func CompositeStorageBucketPrefix(address []byte) ([]byte, bool) {
	return append([]byte(bucketTreeIdentifier), address...), true
}
// construct bucket tree configuration
func SetupBucketConfig(size int, levelGroup int) map[string]interface{} {
	ret := make(map[string]interface{})
	ret[ConfigNumBuckets] = size
	ret[ConfigMaxGroupingAtEachLevel] = levelGroup
	return ret
}

