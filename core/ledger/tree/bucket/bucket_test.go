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
	"reflect"
	"testing"
)

func TestBucket_EncodeDecode(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	bucket := Bucket{
		newDataEntry(newEntryKey("prefix", "key1"), []byte("value1")),
		newDataEntry(newEntryKey("prefix", "key2"), []byte("value2")),
	}
	expect := []byte{2, 0, 0, 0, 0, 0, 0, 0, 4, 107, 101, 121, 49, 6, 118, 97, 108, 117, 101, 49, 4, 107, 101, 121, 50, 6, 118, 97, 108, 117, 101, 50}
	if bytes.Compare(bucket.Encode(), expect) != 0 {
		t.Errorf("expect %v after encode, actual %v", expect, bucket.Encode())
	}

	bucket2 := &Bucket{}
	DecodeBucket("prefix", &Position{5, 7219}, expect, bucket2)
	if !reflect.DeepEqual(*bucket2, bucket) {
		t.Error("expect to be same after decode bucket")
	}
}

func TestBucket_Merge(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	bucket1 := Bucket{
		newDataEntry(newEntryKey("prefix", "key1"), []byte("value1")),
		newDataEntry(newEntryKey("prefix", "key2"), []byte("value2")),
	}
	bucket2 := Bucket{
		newDataEntry(newEntryKey("prefix", "key2"), []byte("newvalue2")),
		newDataEntry(newEntryKey("prefix", "key3"), []byte("value3")),
	}
	bucket3 := bucket1.merge(bucket2)
	expect := Bucket{
		newDataEntry(newEntryKey("prefix", "key1"), []byte("value1")),
		newDataEntry(newEntryKey("prefix", "key2"), []byte("newvalue2")),
		newDataEntry(newEntryKey("prefix", "key3"), []byte("value3")),
	}
	if !reflect.DeepEqual(bucket3, expect) {
		t.Error("expect to be same after bucket merge")
	}
}

func TestBucket_ComputeCrypto(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	bucket := Bucket{
		newDataEntry(newEntryKey("prefix", "key1"), []byte("value1")),
		newDataEntry(newEntryKey("prefix", "key2"), []byte("value2")),
	}
	expect := []byte{139, 71, 255, 217, 168}
	if bytes.Compare(bucket.computeCryptoHash(), expect) != 0 {
		t.Error("expect to be same for bucket compute crypto")
	}
}
