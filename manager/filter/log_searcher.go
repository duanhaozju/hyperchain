package filter

import (
	"github.com/willf/bloom"
	"hyperchain/common"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
)

type Searcher struct {
	addresses  []common.Address
	topics     [][]common.Hash
	begin, end uint64
	namespace  string
}

func NewLogSearcher(begin, end uint64, address []common.Address, topics [][]common.Hash, ns string) *Searcher {
	return &Searcher{
		addresses: address,
		topics:    topics,
		begin:     begin,
		end:       end,
		namespace: ns,
	}
}

func (searcher *Searcher) Search() (ret []*types.Log) {
	var i uint64
	for i = searcher.begin; i <= searcher.end; i += 1 {
		block, err := edb.GetBlockByNumber(searcher.namespace, i)
		if err != nil {
			//
			continue
		}
		if searcher.BloomFilterLookup(block.Bloom) {
			var unfiltered []*types.Log
			for _, tx := range block.Transactions {
				receipt := edb.GetRawReceipt(searcher.namespace, tx.Hash())
				logs, err := receipt.RetrieveLogs()
				if err != nil || len(logs) == 0 {
					continue
				}
				unfiltered = append(unfiltered, logs...)
			}
			ret = append(ret, filterLogs(unfiltered, searcher.criteria())...)
		}
	}
	return
}

func (searcher *Searcher) criteria() *FilterCriteria {
	return &FilterCriteria{
		Addresses: searcher.addresses,
		Topics:    searcher.topics,
	}
}

func (searcher *Searcher) BloomFilterLookup(buf []byte) bool {
	bloom := bloom.New(256*8, 3)
	bloom.GobDecode(buf)

	if len(searcher.addresses) > 0 {
		for _, addr := range searcher.addresses {
			if bloom.Test(addr.Bytes()) {
				return true
			}
		}
		return false
	}

	for _, sub := range searcher.topics {
		var included bool
		for _, topic := range sub {
			if (topic == common.Hash{}) || bloom.Test(topic.Bytes()) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	return true
}
