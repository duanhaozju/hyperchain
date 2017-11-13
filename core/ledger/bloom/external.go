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

package bloom

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
)

func LookupTransaction(namespace string, txHash common.Hash) (bool, error) {
	if bloomFilterCache == nil {
		return false, ErrBloomCacheFailure
	}
	return bloomFilterCache.Look(namespace, txHash)
}

func WriteTxBloomFilter(namespace string, txs []*types.Transaction) (bool, error) {
	if bloomFilterCache == nil {
		return false, ErrBloomCacheFailure
	}
	return bloomFilterCache.Write(namespace, txs)
}

type BloomFilterEvent struct {
	Namespace    string               `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
	Transactions []*types.Transaction `protobuf:"bytes,2,rep,name=Transactions" json:"Transactions,omitempty"`
}

//func WriteTxBloomFilterByEvent(ev event.BloomEvent) {
//	if bloomFilterCache == nil {
//		return false, ErrBloomCacheFailure
//	}
//	bol,err := bloomFilterCache.Write(ev.Namespace, ev.Transactions)
//	if !bol {
//		ev.Cont <- errors.New("Write failed!")
//	}
//	if err != nil {
//		ev.Cont <- err
//	}
//	close(ev.Cont)
//}
