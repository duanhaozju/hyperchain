package hyperstate

func CompositeStorageKey(address []byte, key []byte) []byte {
	return append(address, key...)
}

func CompositeCodeHash(address []byte, codeHash []byte) []byte {
	return  append(address, codeHash)
}
