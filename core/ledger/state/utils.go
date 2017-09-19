package state

import (
	"bytes"
	"hyperchain/common"
	"strconv"
	"time"
)

const (
	storageIdentifier    = "-storage"
	accountIdentifier    = "-account"
	codeIdentifier       = "-code"
	bucketTreeIdentifier = "-bucket"
	journalIdentifier    = "-journal"
)

// ConfigNumBuckets - config name 'numBuckets' as it appears in yaml file
const ConfigNumBuckets = "capacity"

// ConfigMaxGroupingAtEachLevel - config name 'maxGroupingAtEachLevel' as it appears in yaml file
const ConfigMaxGroupingAtEachLevel = "aggreation"

// ConfigBucketCacheMaxSize - config name 'bucketCacheMaxSize' as it appears in yaml file
const ConfigBucketCacheMaxSize = "merkleNodeCache"

// ConfigDataNodeCacheMaxSize - config name 'dataNodeCacheMaxSize' as it appears in yaml file
const ConfigDataNodeCacheMaxSize = "bucketCache"

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

func RetrieveAddrFromStorageKey(key []byte) ([]byte, bool) {
	// Address is const 20byte length
	prefix := []byte(storageIdentifier)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) > prefixLen+common.AddressLength {
		return key[prefixLen : prefixLen+common.AddressLength], true
	} else {
		return nil, false
	}
}

func CompositeArchieveStorageKey(address []byte, key []byte) []byte {
	date := time.Now().Format("20060102")
	ret := append([]byte(storageIdentifier), address...)
	ret = append(ret, []byte(date)...)
	ret = append(ret, key...)
	return ret
}

func GetArchieveStorageKeyPrefix(address []byte) []byte {
	ret := append([]byte(storageIdentifier), address...)
	return ret
}

func GetArchieveStorageKeyWithDatePrefix(address []byte, date []byte) []byte {
	ret := append([]byte(storageIdentifier), address...)
	ret = append(ret, date...)
	return ret
}

func SplitCompositeArchieveStorageKey(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(storageIdentifier), address...)
	prefixLen := len(prefix) + 8
	if bytes.HasPrefix(key, prefix) && len(key) >= prefixLen {
		return key[prefixLen:], true
	} else {
		return nil, false
	}
}

func GetArchieveDate(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(storageIdentifier), address...)
	prefixLen := len(prefix) + 8
	if bytes.HasPrefix(key, prefix) && len(key) >= prefixLen {
		return key[len(prefix):prefixLen], true
	} else {
		return nil, false
	}

}

/*
	Code
*/
func CompositeCodeHash(address []byte, codeHash []byte) []byte {
	ret := append([]byte(codeIdentifier), address...)
	return append(ret, codeHash...)
}

func SplitCompositeCodeHash(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(codeIdentifier), address...)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) >= prefixLen {
		return key[prefixLen:], true
	} else {
		return nil, false
	}
}

func RetrieveAddrFromCodeHash(key []byte) ([]byte, bool) {
	// Address is const 20byte length
	prefix := []byte(codeIdentifier)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) > prefixLen+common.AddressLength {
		return key[prefixLen : prefixLen+common.AddressLength], true
	} else {
		return nil, false
	}
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
func CompositeStorageBucketPrefix(address string) ([]byte, bool) {
	return []byte(bucketTreeIdentifier + address), true
}

// construct bucket tree configuration
func SetupBucketConfig(size, levelGroup, bucketCacheMaxSize, dataNodeCacheMaxSize int) map[string]interface{} {
	ret := make(map[string]interface{})
	ret[ConfigNumBuckets] = size
	ret[ConfigMaxGroupingAtEachLevel] = levelGroup
	ret[ConfigBucketCacheMaxSize] = bucketCacheMaxSize
	ret[ConfigDataNodeCacheMaxSize] = dataNodeCacheMaxSize
	return ret
}

/*
	Journal
*/

func CompositeJournalKey(blockNumber uint64) []byte {
	s := strconv.FormatUint(blockNumber, 10)
	return append([]byte(journalIdentifier), []byte(s)...)
}
