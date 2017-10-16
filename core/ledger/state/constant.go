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
package state

const (
	// Normal account status. Account is allowed to be invoked.
	OBJ_NORMAL = iota
	// Frozon account status. Account is not allowed to be invoked.
	OBJ_FROZON
)

const (
	BATCH_NORMAL = iota
	BATCH_ARCHIVE
)

const (
	// Special opcode. When opcode is 100, all deleted storage content will be saved into historical db.
	// This kind of operation is regarded as a archive operation.
	opcodeArchive = 100
	// whether turn on fake hash function
	enableFakeHashFn = false
	// whether to remove empty stateObject
	deleteEmptyObjects = true
)
