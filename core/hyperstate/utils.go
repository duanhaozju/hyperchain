package hyperstate

import "bytes"

const storageIdentifier = "Storage"

func CompositeStorageKey(address []byte, key []byte) []byte {
	ret := append(address, []byte(storageIdentifier)...)
	ret = append(ret, key...)
	return ret
}

func GetStorageKeyPrefix(address []byte) []byte {
	ret := append(address, []byte(storageIdentifier)...)
	return ret
}

func SplitCompositeStorageKey(address []byte, key []byte) []byte {
	flag := append(address, []byte(storageIdentifier)...)
	idx := bytes.Index(key, flag)
	if idx == - 1 {
		return nil
	}
	return key[idx:]
}

func CompositeCodeHash(address []byte, codeHash []byte) []byte {
	return  append(address, codeHash...)
}
