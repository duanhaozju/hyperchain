package bloom

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"github.com/pkg/errors"
	"hyperchain/manager/event"
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

func WriteTxBloomFilterByEvent(ev event.BloomEvent) {
	if bloomFilterCache == nil {
		return false, ErrBloomCacheFailure
	}
	bol,err := bloomFilterCache.Write(ev.Namespace, ev.Transactions)
	if !bol {
		ev.Cont <- errors.New("Write failed!")
	}
	if err != nil {
		ev.Cont <- err
	}
	close(ev.Cont)
}
