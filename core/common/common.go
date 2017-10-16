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
package common

import (
	cm "hyperchain/common"
	"hyperchain/hyperdb"
	"path"
)

var precompiledAccount = map[string]struct{}{
	cm.BytesToAddress(cm.LeftPadBytes([]byte{1}, 20)).Hex(): struct{}{}, // ECRECOVER
	cm.BytesToAddress(cm.LeftPadBytes([]byte{2}, 20)).Hex(): struct{}{}, // SHA256
	cm.BytesToAddress(cm.LeftPadBytes([]byte{3}, 20)).Hex(): struct{}{}, // RIPEMD160
	cm.BytesToAddress(cm.LeftPadBytes([]byte{4}, 20)).Hex(): struct{}{}, // MEMCPY
}

func RetrieveSnapshotFileds() []string {
	return []string{
		// world state related
		"-account",
		"-storage",
		"-code",
		// bucket tree related
		"-bucket",
		"BucketNode",
	}
}

func IsPrecompiledAccount(address cm.Address) bool {
	if _, exist := precompiledAccount[address.Hex()]; exist {
		return true
	}
	return false
}

func GetDatabaseHome(namespace string, conf *cm.Config) string {
	return path.Join("namespaces", namespace, hyperdb.GetDatabaseHome(conf))
}
