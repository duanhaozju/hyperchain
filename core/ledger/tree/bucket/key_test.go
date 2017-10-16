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
	"fmt"
	"github.com/ToQoz/gopwt/assert"
	"reflect"
	"testing"
)

func NewTestPosition() *Position {
	var (
		l = rnd.Intn(conf.lowest + 1)
		i = rnd.Intn(conf.getMaxIndex(l)) + 1
	)
	return &Position{l, i}
}

func TestPositionGetParent(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	poses := []*Position{{1, 1}, {1, 10}, {5, 1}, {5, 10009}}
	parents := []*Position{{0, 1}, {0, 1}, {4, 1}, {4, 1001}}
	for idx, pos := range poses {
		parent := pos.getParent()
		assert.OK(t, reflect.DeepEqual(parents[idx], parent), "parent not equal")
	}
}

func TestPostionEqual(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	poses := []*Position{{1, 1}, {1, 10}, {5, 1}, {5, 10009}}
	comp := []*Position{{1, 1}, {1, 10}, {5, 1}, {5, 10009}}
	for idx, pos := range poses {
		assert.OK(t, pos.equals(comp[idx]), "two position should be equal")
	}
}

func TestPositionGetChildKey(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	poses := []*Position{{1, 1}, {2, 10}}
	children := []*Position{{2, 1}, {2, 10}, {3, 91}, {3, 100}}
	for idx, pos := range poses {
		child := pos.getChildKey(1)
		assert.OK(t, child.level == children[idx*2].level)
		assert.OK(t, child.index == children[idx*2].index)
		child = pos.getChildKey(10)
		assert.OK(t, child.level == children[idx*2+1].level)
		assert.OK(t, child.index == children[idx*2+1].index)
	}
}

func TestPositionGetChildIndex(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	indexes := []int{1, 2, 10}
	children := []*Position{{5, 1}, {5, 2}, {5, 10}}
	pos := &Position{4, 1}
	for i, idx := range indexes {
		assert.OK(t, reflect.DeepEqual(pos.getChildKey(idx), children[i]), "child position should be same")
	}
}

func TestPositionEncode(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	pos := &Position{4, 1}
	assert.OK(t, bytes.Compare(pos.encode(), []byte{0, 4, 1}) == 0, "encoded value should be same")
}

func TestNewEntryKey(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	k := newEntryKey(prefix, "key")
	fmt.Println(k.pos)
	fmt.Println(k.key)
	assert.OK(t, reflect.DeepEqual(&Position{5, 1131}, k.pos), "postion should be same")
	assert.OK(t, bytes.Compare(k.key, []byte{116, 114, 101, 101, 80, 114, 101, 102, 105, 120, 0, 107, 101, 121}) == 0, "composite key should be same")
}
