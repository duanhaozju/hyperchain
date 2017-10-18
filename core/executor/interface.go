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

package executor

// A non-verifying peer is a function of a clipped block chain node.
// Such nodes do not participate in the consensus, only participate in blockchain synchronization.
// The non-verifying node is 100 percent trusted by the connected verifying node,
// In another word, nvp will not check block authenticity received from the vp.
//
// Non-verifying peer can relay transaction to verifying peer and generate a transaction hash to client.
// Besides, nvp also provide blockchain data query and transaction simulation.
type NonVerifyingPeer interface {
	ReceiveBlock([]byte)
}
