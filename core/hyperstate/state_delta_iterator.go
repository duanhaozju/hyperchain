package hyperstate

// StateDeltaIterator - An iterator implementation over state-delta
type StateDeltaIterator struct {
	updates         map[string]*UpdatedValue
	relevantKeys    []string
	currentKeyIndex int
	done            bool
}

// NewStateDeltaRangeScanIterator - return an iterator for performing a range scan over a state-delta object
func NewStateDeltaRangeScanIterator(delta *StateDelta, accountID string, startKey string, endKey string) *StateDeltaIterator {
	updates := delta.GetUpdates(accountID)
	return &StateDeltaIterator{updates, retrieveRelevantKeys(updates, startKey, endKey), -1, false}
}

func retrieveRelevantKeys(updates map[string]*UpdatedValue, startKey string, endKey string) []string {
	relevantKeys := []string{}
	if updates == nil {
		return relevantKeys
	}
	for k, v := range updates {
		if k >= startKey && (endKey == "" || k <= endKey) && !v.IsDeleted() {
			relevantKeys = append(relevantKeys, k)
		}
	}
	return relevantKeys
}

// Next - see interface 'RangeScanIterator' for details
func (itr *StateDeltaIterator) Next() bool {
	itr.currentKeyIndex++
	if itr.currentKeyIndex < len(itr.relevantKeys) {
		return true
	}
	itr.currentKeyIndex--
	itr.done = true
	return false
}

// GetKeyValue - see interface 'RangeScanIterator' for details
func (itr *StateDeltaIterator) GetKeyValue() (string, []byte) {
	if itr.done {
		log.Warning("Iterator used after it has been exhausted. Last retrieved value will be returned")
	}
	key := itr.relevantKeys[itr.currentKeyIndex]
	value := itr.updates[key].GetValue()
	return key, value
}

// Close - see interface 'RangeScanIterator' for details
func (itr *StateDeltaIterator) Close() {
}

// ContainsKey - checks wether the given key is present in the state-delta
func (itr *StateDeltaIterator) ContainsKey(key string) bool {
	_, ok := itr.updates[key]
	return ok
}
