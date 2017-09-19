// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bucket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gogo/protobuf/proto"
)

var stateKeyDelimiter = []byte{0x00}

// Append the [][]byte with separator,it will be more efficient in high concurrency
func JoinBytes(data [][]byte, separator string) []byte {
	// count the length of separator
	if data == nil || len(data) == 0 {
		return nil
	}
	n := len(separator) * (len(data) - 1)
	for i := 0; i < len(data); i++ {
		n = n + len(data[i])
	}
	// allocate the memory of data
	mem := make([]byte, n)

	// copy data
	dp := copy(mem, data[0])
	for _, bytesData := range data[1:] {
		dp += copy(mem[dp:], separator)
		dp += copy(mem[dp:], bytesData)
	}
	return mem
}

// EncodeOrderPreservingVarUint64 returns a byte-representation for a uint64 number such that
// all zero-bits starting bytes are trimmed in order to reduce the length of the array
// For preserving the order in a default bytes-comparison, first byte contains the number of remaining bytes.
// The presence of first byte also allows to use the returned bytes as part of other larger byte array such as a
// composite-key representation in db
func EncodeOrderPreservingVarUint64(number uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, number)
	startingIndex := 0
	size := 0
	for i, b := range bytes {
		if b != 0x00 {
			startingIndex = i
			size = 8 - i
			break
		}
	}
	sizeBytes := proto.EncodeVarint(uint64(size))
	if len(sizeBytes) > 1 {
		panic(fmt.Errorf("[]sizeBytes should not be more than one byte because the max number it needs to hold is 8. size=%d", size))
	}
	encodedBytes := make([]byte, size+1)
	encodedBytes[0] = sizeBytes[0]
	copy(encodedBytes[1:], bytes[startingIndex:])
	return encodedBytes
}

// DecodeOrderPreservingVarUint64 decodes the number from the bytes obtained from method 'EncodeOrderPreservingVarUint64'.
// Also, returns the number of bytes that are consumed in the process
func DecodeOrderPreservingVarUint64(bytes []byte) (uint64, int) {
	s, _ := proto.DecodeVarint(bytes)
	size := int(s)
	decodedBytes := make([]byte, 8)
	copy(decodedBytes[8-size:], bytes[1:size+1])
	numBytesConsumed := size + 1
	return binary.BigEndian.Uint64(decodedBytes), numBytesConsumed
}

// ConstructCompositeKey returns a []byte that uniquely represents a given prefix and key.
// This assumes that prefix does not contain a 0x00 byte, but the key may
func ConstructCompositeKey(prefix string, key string) []byte {
	return bytes.Join([][]byte{[]byte(prefix), []byte(key)}, stateKeyDelimiter)
}

// DecodeCompositeKey decodes the compositeKey constructed by ConstructCompositeKey method
// back to the original prefix and key form
func DecodeCompositeKey(compositeKey []byte) (string, string) {
	split := bytes.SplitN(compositeKey, stateKeyDelimiter, 2)
	return string(split[0]), string(split[1])
}
