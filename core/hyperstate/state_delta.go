package hyperstate

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"sort"
)

/*
	StateDelta
*/
// StateDelta holds the changes to existing state. This struct is used for holding the uncommitted changes during execution of a tx-batch
// Also, to be used for transferring the state to another peer in chunks
type StateDelta struct {
	AccountDeltas map[string]*AccountDelta
	// RollBackwards allows one to contol whether this delta will roll the state
	// forwards or backwards.
	RollBackwards bool
}

// NewStateDelta constructs an empty StateDelta struct
func NewStateDelta() *StateDelta {
	return &StateDelta{make(map[string]*AccountDelta), false}
}

// Get get the state from delta if exists
func (stateDelta *StateDelta) Get(accountID string, key string) *UpdatedValue {
	// TODO Cache?
	accountStateDelta, ok := stateDelta.AccountDeltas[accountID]
	if ok {
		return accountStateDelta.get(key)
	}
	return nil
}

// Set sets state value for a key
func (stateDelta *StateDelta) Set(accountID string, key string, value, previousValue []byte) {
	accountStateDelta := stateDelta.getOrCreateAccountDelta(accountID)
	accountStateDelta.set(key, value, previousValue)
	return
}

// Delete deletes a key from the state
func (stateDelta *StateDelta) Delete(accountID string, key string, previousValue []byte) {
	accountStateDelta := stateDelta.getOrCreateAccountDelta(accountID)
	accountStateDelta.remove(key, previousValue)
	return
}

// IsUpdatedValueSet returns true if a update value is already set for
// the given account ID and key.
func (stateDelta *StateDelta) IsUpdatedValueSet(accountID, key string) bool {
	accountStateDelta, ok := stateDelta.AccountDeltas[accountID]
	if !ok {
		return false
	}
	if _, ok := accountStateDelta.UpdatedKVs[key]; ok {
		return true
	}
	return false
}

// ApplyChanges merges another delta - if a key is present in both, the value of the existing key is overwritten
func (stateDelta *StateDelta) ApplyChanges(anotherStateDelta *StateDelta) {
	for accountID, accountStateDelta := range anotherStateDelta.AccountDeltas {
		existingAccountDelta, existed := stateDelta.AccountDeltas[accountID]
		for key, valueHolder := range accountStateDelta.UpdatedKVs {
			var previousValue []byte
			if existed {
				existingUpdateValue, existingUpdate := existingAccountDelta.UpdatedKVs[key]
				if existingUpdate {
					// The existing state delta already has an updated value for this key.
					previousValue = existingUpdateValue.PreviousValue
				} else {
					// Use the previous value set in the new state delta
					previousValue = valueHolder.PreviousValue
				}
			} else {
				// Use the previous value set in the new state delta
				previousValue = valueHolder.PreviousValue
			}

			if valueHolder.IsDeleted() {
				stateDelta.Delete(accountID, key, previousValue)
			} else {
				stateDelta.Set(accountID, key, valueHolder.Value, previousValue)
			}
		}
	}
}

// IsEmpty checks whether StateDelta contains any data
func (stateDelta *StateDelta) IsEmpty() bool {
	return len(stateDelta.AccountDeltas) == 0
}

// GetUpdatedAccountIds return the accountIDs that are prepsent in the delta
// If sorted is true, the method return accountIDs in lexicographical sorted order
func (stateDelta *StateDelta) GetUpdatedAccountIds(sorted bool) []string {
	updatedAccountIds := make([]string, len(stateDelta.AccountDeltas))
	i := 0
	for k := range stateDelta.AccountDeltas {
		updatedAccountIds[i] = k
		i++
	}
	if sorted {
		sort.Strings(updatedAccountIds)
	}
	return updatedAccountIds
}

// GetUpdates returns changes associated with given chaincodeId
func (stateDelta *StateDelta) GetUpdates(accountID string) map[string]*UpdatedValue {
	accountStateDelta := stateDelta.AccountDeltas[accountID]
	if accountStateDelta == nil {
		return nil
	}
	return accountStateDelta.UpdatedKVs
}

func (stateDelta *StateDelta) getOrCreateAccountDelta(accountID string) *AccountDelta {
	accountStateDelta, ok := stateDelta.AccountDeltas[accountID]
	if !ok {
		accountStateDelta = newAccountDelta(accountID)
		stateDelta.AccountDeltas[accountID] = accountStateDelta
	}
	return accountStateDelta
}

// ComputeCryptoHash computes crypto-hash for the data held
// returns nil if no data is present
func (stateDelta *StateDelta) ComputeCryptoHash() []byte {
	if stateDelta.IsEmpty() {
		return nil
	}
	var buffer bytes.Buffer
	sortedAccountIds := stateDelta.GetUpdatedAccountIds(true)
	for _, accountID := range sortedAccountIds {
		buffer.WriteString(accountID)
		accountStateDelta := stateDelta.AccountDeltas[accountID]
		sortedKeys := accountStateDelta.getSortedKeys()
		for _, key := range sortedKeys {
			buffer.WriteString(key)
			updatedValue := accountStateDelta.get(key)
			if !updatedValue.IsDeleted() {
				buffer.Write(updatedValue.Value)
			}
		}
	}
	hashingContent := buffer.Bytes()
	return kec256Hash.Hash(hashingContent).Bytes()
}

/*
	AccountDelta
*/
//AccountDelta maintains state for a chaincode
type AccountDelta struct {
	AccountID  string
	UpdatedKVs map[string]*UpdatedValue
}

func newAccountDelta(accountID string) *AccountDelta {
	return &AccountDelta{accountID, make(map[string]*UpdatedValue)}
}

func (accountStateDelta *AccountDelta) get(key string) *UpdatedValue {
	// TODO Cache?
	return accountStateDelta.UpdatedKVs[key]
}

func (accountStateDelta *AccountDelta) set(key string, updatedValue, previousValue []byte) {
	updatedKV, ok := accountStateDelta.UpdatedKVs[key]
	if ok {
		// Key already exists, just set the updated value
		updatedKV.Value = updatedValue
	} else {
		// New key. Create a new entry in the map
		accountStateDelta.UpdatedKVs[key] = &UpdatedValue{updatedValue, previousValue}
	}
}

func (accountStateDelta *AccountDelta) remove(key string, previousValue []byte) {
	updatedKV, ok := accountStateDelta.UpdatedKVs[key]
	if ok {
		// Key already exists, just set the value
		updatedKV.Value = nil
	} else {
		// New key. Create a new entry in the map
		accountStateDelta.UpdatedKVs[key] = &UpdatedValue{nil, previousValue}
	}
}

func (accountStateDelta *AccountDelta) hasChanges() bool {
	return len(accountStateDelta.UpdatedKVs) > 0
}

func (accountStateDelta *AccountDelta) getSortedKeys() []string {
	updatedKeys := []string{}
	for k := range accountStateDelta.UpdatedKVs {
		updatedKeys = append(updatedKeys, k)
	}
	sort.Strings(updatedKeys)
	return updatedKeys
}

/*
	UpdatedValue
*/
// UpdatedValue holds the value for a key
type UpdatedValue struct {
	Value         []byte
	PreviousValue []byte
}

// IsDeleted checks whether the key was deleted
func (updatedValue *UpdatedValue) IsDeleted() bool {
	return updatedValue.Value == nil
}

// GetValue returns the value
func (updatedValue *UpdatedValue) GetValue() []byte {
	return updatedValue.Value
}

// GetPreviousValue returns the previous value
func (updatedValue *UpdatedValue) GetPreviousValue() []byte {
	return updatedValue.PreviousValue
}

/*
	Marshal
*/
// marshalling / Unmarshalling code
// We need to revisit the following when we define proto messages
// for state related structures for transporting. May be we can
// completely get rid of custom marshalling / Unmarshalling of a state delta

// Marshal serializes the StateDelta
func (stateDelta *StateDelta) Marshal() (b []byte) {
	buffer := proto.NewBuffer([]byte{})
	err := buffer.EncodeVarint(uint64(len(stateDelta.AccountDeltas)))
	if err != nil {
		// in protobuf code the error return is always nil
		panic(fmt.Errorf("This error should not occure: %s", err))
	}
	for accountID, accountStateDelta := range stateDelta.AccountDeltas {
		buffer.EncodeStringBytes(accountID)
		accountStateDelta.marshal(buffer)
	}
	b = buffer.Bytes()
	return
}

func (accountStateDelta *AccountDelta) marshal(buffer *proto.Buffer) {
	err := buffer.EncodeVarint(uint64(len(accountStateDelta.UpdatedKVs)))
	if err != nil {
		panic(fmt.Errorf("This error should not occur: %s", err))
	}
	for key, valueHolder := range accountStateDelta.UpdatedKVs {
		err = buffer.EncodeStringBytes(key)
		if err != nil {
			panic(fmt.Errorf("This error should not occur: %s", err))
		}
		accountStateDelta.marshalValueWithMarker(buffer, valueHolder.Value)
		accountStateDelta.marshalValueWithMarker(buffer, valueHolder.PreviousValue)
	}
	return
}

func (accountStateDelta *AccountDelta) marshalValueWithMarker(buffer *proto.Buffer, value []byte) {
	if value == nil {
		// Just add a marker that the value is nil
		err := buffer.EncodeVarint(uint64(0))
		if err != nil {
			panic(fmt.Errorf("This error should not occur: %s", err))
		}
		return
	}
	err := buffer.EncodeVarint(uint64(1))
	if err != nil {
		panic(fmt.Errorf("This error should not occur: %s", err))
	}
	// If the value happen to be an empty byte array, it would appear as a nil during
	// deserialization - see method 'unmarshalValueWithMarker'
	err = buffer.EncodeRawBytes(value)
	if err != nil {
		panic(fmt.Errorf("This error should not occur: %s", err))
	}
}

/*
	Unmarshal
*/
// Unmarshal deserializes StateDelta
func (stateDelta *StateDelta) Unmarshal(bytes []byte) error {
	buffer := proto.NewBuffer(bytes)
	size, err := buffer.DecodeVarint()
	if err != nil {
		return fmt.Errorf("Error unmarashaling size: %s", err)
	}
	stateDelta.AccountDeltas = make(map[string]*AccountDelta, size)
	for i := uint64(0); i < size; i++ {
		accountID, err := buffer.DecodeStringBytes()
		if err != nil {
			return fmt.Errorf("Error unmarshaling accountID : %s", err)
		}
		accountStateDelta := newAccountDelta(accountID)
		err = accountStateDelta.unmarshal(buffer)
		if err != nil {
			return fmt.Errorf("Error unmarshalling accountStateDelta : %s", err)
		}
		stateDelta.AccountDeltas[accountID] = accountStateDelta
	}

	return nil
}

func (accountStateDelta *AccountDelta) unmarshal(buffer *proto.Buffer) error {
	size, err := buffer.DecodeVarint()
	if err != nil {
		return fmt.Errorf("Error unmarshaling state delta: %s", err)
	}
	accountStateDelta.UpdatedKVs = make(map[string]*UpdatedValue, size)
	for i := uint64(0); i < size; i++ {
		key, err := buffer.DecodeStringBytes()
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		value, err := accountStateDelta.unmarshalValueWithMarker(buffer)
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		previousValue, err := accountStateDelta.unmarshalValueWithMarker(buffer)
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		accountStateDelta.UpdatedKVs[key] = &UpdatedValue{value, previousValue}
	}
	return nil
}

func (accountStateDelta *AccountDelta) unmarshalValueWithMarker(buffer *proto.Buffer) ([]byte, error) {
	valueMarker, err := buffer.DecodeVarint()
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling state delta : %s", err)
	}
	if valueMarker == 0 {
		return nil, nil
	}
	value, err := buffer.DecodeRawBytes(false)
	if err != nil {
		return nil, fmt.Errorf("Error unmarhsaling state delta : %s", err)
	}
	// protobuff makes an empty []byte into a nil. So, assigning an empty byte array explicitly
	if value == nil {
		value = []byte{}
	}
	return value, nil
}
