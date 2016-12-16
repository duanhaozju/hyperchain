package hyperstate

import (
	"testing"
	"hyperchain/common"
	"math/big"
)

func TestMarshal(t *testing.T) {
	var jourList []journalEntry
	addr1 := common.BytesToAddress([]byte("1"))
	addr2 := common.BytesToAddress([]byte("2"))
	addr3 := common.BytesToAddress([]byte("3"))
	addr4 := common.BytesToAddress([]byte("4"))
	addr5 := common.BytesToAddress([]byte("5"))
	addr6 := common.BytesToAddress([]byte("6"))
	addr9 := common.BytesToAddress([]byte("9"))
	jourList = append(jourList, &createObjectChange{
		account: &addr1,
	})
	jourList = append(jourList, &suicideChange{
		account: &addr2,
		prev: true,
		prevbalance: big.NewInt(2),
	})
	jourList = append(jourList, &balanceChange{
		account: &addr3,
		prev:  big.NewInt(3),
	})
	jourList = append(jourList, &nonceChange{
		account: &addr4,
		prev: uint64(4),
	})
	jourList = append(jourList, &storageChange{
		account: &addr5,
		key: common.BytesToHash([]byte("key5")),
		prevalue: common.BytesToHash([]byte("value5")),
	})
	jourList = append(jourList, &codeChange{
		account: &addr6,
		prevcode: []byte("code6"),
		prevhash: []byte("hash6"),
	})
	jourList = append(jourList, &refundChange{
		prev: big.NewInt(7),
	})
	jourList = append(jourList, &addLogChange{
		txhash:  common.BytesToHash([]byte("txhash8")),
	})
	jourList = append(jourList, &touchChange{
		account: &addr9,
		prev: true,
	})

	test := rong{journalList: jourList}
	t.Log(test)
	res, err := test.Marshal()
	t.Log(res)
	if err != nil {
		t.Error("journal marshal err")
	}

	jo := rong{}
	err = UnmarshalJournal(res, &jo)
	t.Log(jo)

}

