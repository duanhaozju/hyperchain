package bucket

import (
	"math/big"
	"github.com/golang/protobuf/proto"
	"fmt"
	"sort"
)
// the value which updated by datanode
type UpdatedValue struct {
	Value         []byte
	PreviousValue []byte
}

// the set of UpdatedValue
type UpdatedValueSet struct {
	BlockNum   *big.Int
	UpdatedKVs map[string] *UpdatedValue
}


func (updatedValueSet *UpdatedValueSet) Marshal(buffer *proto.Buffer){
	err := buffer.EncodeVarint(uint64(len(updatedValueSet.UpdatedKVs)))
	if err != nil {
		panic(fmt.Errorf("This error should not occur: %s", err))
	}
	for key, valueHolder := range updatedValueSet.UpdatedKVs {
		err = buffer.EncodeStringBytes(key)
		if err != nil {
			panic(fmt.Errorf("This error should not occur: %s", err))
		}
		updatedValueSet.marshalValueWithMarker(buffer, valueHolder.Value)
		updatedValueSet.marshalValueWithMarker(buffer, valueHolder.PreviousValue)
	}
	return
}

func (updatedValueSet *UpdatedValueSet) marshalValueWithMarker(buffer *proto.Buffer, value []byte) {
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

func (updatedValueSet *UpdatedValueSet) UnMarshal(buffer *proto.Buffer) error {
	size, err := buffer.DecodeVarint()
	if err != nil {
		return fmt.Errorf("Error unmarshaling state delta: %s", err)
	}
	updatedValueSet.UpdatedKVs = make(map[string]*UpdatedValue, size)
	for i := uint64(0); i < size; i++ {
		key, err := buffer.DecodeStringBytes()
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		value, err := updatedValueSet.unmarshalValueWithMarker(buffer)
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		previousValue, err := updatedValueSet.unmarshalValueWithMarker(buffer)
		if err != nil {
			return fmt.Errorf("Error unmarshaling state delta : %s", err)
		}
		updatedValueSet.UpdatedKVs[key] = &UpdatedValue{value, previousValue}
	}
	return nil
}

func (updatedValueSet *UpdatedValueSet) unmarshalValueWithMarker(buffer *proto.Buffer) ([]byte, error) {
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


func (updatedValueSet *UpdatedValueSet) Get(key string) *UpdatedValue {
	// TODO Cache?
	return updatedValueSet.UpdatedKVs[key]
}

func (updatedValueSet *UpdatedValueSet) Set(key string, updatedValue, previousValue []byte) {
	updatedKVs, ok := updatedValueSet.UpdatedKVs[key]
	if ok {
		// Key already exists, just set the updated value
		updatedKVs.Value = updatedValue
	} else {
		// New key. Create a new entry in the map
		updatedValueSet.UpdatedKVs[key] = &UpdatedValue{updatedValue, previousValue}
	}
}

func (updatedValueSet *UpdatedValueSet) Remove(key string, previousValue []byte) {
	updatedKVs, ok := updatedValueSet.UpdatedKVs[key]
	if ok {
		// Key already exists, just set the value
		updatedKVs.Value = nil
	} else {
		// New key. Create a new entry in the map
		updatedValueSet.UpdatedKVs[key] = &UpdatedValue{nil, previousValue}
	}
}

func (updatedValueSet *UpdatedValueSet) HasChanges() bool {
	return len(updatedValueSet.UpdatedKVs) > 0
}

func (updatedValueSet *UpdatedValueSet) getSortedKeys() []string {
	updatedKeys := []string{}
	for k := range updatedValueSet.UpdatedKVs {
		updatedKeys = append(updatedKeys, k)
	}
	sort.Strings(updatedKeys)
	logger.Debugf("Sorted keys = %#v", updatedKeys)
	return updatedKeys
}