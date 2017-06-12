package executor

var RemoveLessThan = func(key interface{}, iterKey interface{}) bool {
	id := key.(uint64)
	iterId := iterKey.(uint64)
	if id >= iterId {
		return true
	}
	return false
}
