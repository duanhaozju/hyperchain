package hyperstate

import (
	"testing"
	"hyperchain/common"
	"math/big"
)

func TestMarshal(t *testing.T) {
	var jourList []JournalEntry
	addr1 := common.BytesToAddress([]byte("1"))
	addr2 := common.BytesToAddress([]byte("2"))
	addr3 := common.BytesToAddress([]byte("3"))
	addr4 := common.BytesToAddress([]byte("4"))
	addr5 := common.BytesToAddress([]byte("5"))
	addr6 := common.BytesToAddress([]byte("6"))
	addr9 := common.BytesToAddress([]byte("9"))
	jourList = append(jourList, &CreateObjectChange{
		Account: &addr1,
	})
	jourList = append(jourList, &SuicideChange{
		Account: &addr2,
		Prev: true,
		Prevbalance: big.NewInt(2),
	})
	jourList = append(jourList, &BalanceChange{
		BalanceAccount: &addr3,
		BalancePrev:  big.NewInt(3),
	})
	jourList = append(jourList, &NonceChange{
		NonceAccount: &addr4,
		NoncePrev: uint64(4),
	})
	jourList = append(jourList, &StorageChange{
		Account: &addr5,
		Key: common.BytesToHash([]byte("key5")),
		Prevalue: common.BytesToHash([]byte("value5")),
	})
	jourList = append(jourList, &CodeChange{
		Account: &addr6,
		Prevcode: []byte("code6"),
		Prevhash: []byte("hash6"),
	})
	jourList = append(jourList, &RefundChange{
		Prev: big.NewInt(7),
	})
	jourList = append(jourList, &AddLogChange{
		Txhash:  common.BytesToHash([]byte("txhash8")),
	})
	jourList = append(jourList, &TouchChange{
		TouchAccount: &addr9,
		TouchPrev: true,
	})

	test := &Journal{JournalList: jourList}
	t.Log("test: ", test)
	res, err := test.Marshal()
	if err != nil {
		t.Error("journal marshal err")
	}
	jo, err := UnmarshalJournal(res)
	t.Log("err: ", err, "jo: ", jo)
}