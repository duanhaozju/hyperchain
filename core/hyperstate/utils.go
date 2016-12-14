package hyperstate

import "bytes"

const storageIdentifier = "-Storage"
const accountIdentifier = "-Account"
const codeIdentifier = "-Code"

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

func SplitCompositeStorageKey(address []byte, key []byte) []byte {
	prefix := append([]byte(storageIdentifier), address...)
	idx := bytes.Index(key, prefix)
	if idx == - 1 {
		return nil
	}
	return key[idx:]
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

func SplitCompositeAccountKey(key []byte) []byte {
	prefix := []byte(accountIdentifier)
	idx := bytes.Index(key, prefix)
	if idx == - 1 {
		return nil
	}
	return key[idx:]
}
