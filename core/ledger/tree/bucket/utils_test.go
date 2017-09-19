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
	"testing"
)

func TestJoinBytes(t *testing.T) {
	type testcase struct {
		input  [][]byte
		expect []byte
	}
	cases := []testcase{
		{nil, nil},
		{[][]byte{[]byte{0, 1}, []byte{2, 3}}, []byte{0, 1, 2, 3}},
	}
	for _, c := range cases {
		if bytes.Compare(c.expect, JoinBytes(c.input, "")) != 0 {
			t.Errorf("expect to be same, input %v, expect %v, acutal %v", c.input, c.expect, JoinBytes(c.input, ""))
		}
	}
}

func TestEncodeOrderPreservingVarUint64(t *testing.T) {
	type testcase struct {
		input  uint64
		expect []byte
	}
	cases := []testcase{
		{0, []byte{0x00}},
		{1, []byte{0x01, 0x01}},
		{65535, []byte{0x02, 0xff, 0xff}},
	}
	for _, c := range cases {
		if bytes.Compare(c.expect, EncodeOrderPreservingVarUint64(c.input)) != 0 {
			t.Errorf("expect to be same, input %v, expect %v, acutal %v", c.input, c.expect, EncodeOrderPreservingVarUint64(c.input))
		}
	}
}

func TestDecodeOrderPreservingVarUint64(t *testing.T) {
	type testcase struct {
		input  []byte
		expect uint64
	}
	cases := []testcase{
		{[]byte{0x00}, 0},
		{[]byte{0x01, 0x01}, 1},
		{[]byte{0x02, 0xff, 0xff}, 65535},
	}
	for _, c := range cases {
		res, _ := DecodeOrderPreservingVarUint64(c.input)
		if c.expect != res {
			t.Errorf("expect to be same, input %v, expect %v, acutal %v", c.input, c.expect, res)
		}
	}
}

func TestConstructCompositeKey(t *testing.T) {
	type testcase struct {
		prefix string
		key    string
		expect []byte
	}
	cases := []testcase{
		{"", "", []byte{0x00}},
		{string([]byte{0x01, 0x02}), string([]byte{0x03, 0x04}), []byte{0x01, 0x02, 0x00, 0x03, 0x04}},
	}
	for _, c := range cases {
		if bytes.Compare(ConstructCompositeKey(c.prefix, c.key), c.expect) != 0 {
			t.Errorf("expect to be same with expect, prefix %v, key %v, expect %v, actual %v",
				c.prefix, c.key, c.expect, ConstructCompositeKey(c.prefix, c.key))
		}
	}
}

func TestDecodeCompositeKey(t *testing.T) {
	type testcase struct {
		prefix string
		key    string
		input  []byte
	}
	cases := []testcase{
		{"", "", []byte{0x00}},
		{string([]byte{0x01, 0x02}), string([]byte{0x03, 0x04}), []byte{0x01, 0x02, 0x00, 0x03, 0x04}},
	}
	for _, c := range cases {
		prefix, key := DecodeCompositeKey(c.input)
		if prefix != c.prefix || key != c.key {
			t.Errorf("expect to be same with expect, input %v, expect prefix %s, acutal prefix %s, expect key %s, actual key %s",
				c.input, c.prefix, prefix, c.key, key)
		}
	}
}
