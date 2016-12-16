package hyperstate

import (
	"testing"
	"hyperchain/common"
	"math/big"
	"encoding/json"
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

func TestPointer(t *testing.T) {
	type FamilyMember struct {
		Name    string
		Age     int
		Parents []string
	}
	type A struct {
		Obj FamilyMember
	}
	type B struct {
		Obj *FamilyMember
	}

	fam := FamilyMember{
		Name: "hfl",
		Age: 18,
		Parents: []string{"mike", "alice"},
	}

	t.Log("fam: ", fam)
	tmp1 := &A{Obj: fam}
	tmp2 := &B{Obj: &fam}

	t.Log("tmp1: ", tmp1, "tmp2: ", tmp2)

	res1, err1 := json.Marshal(tmp1)
	res2, err2 := json.Marshal(tmp2)
	t.Log("res1: ", string(res1), "err1: ", err1)
	t.Log("res2: ", string(res2), "err2: ", err2)

	tmp3 := A{}
	tmp4 := B{}
	err3 := json.Unmarshal(res1, &tmp3)
	err4 := json.Unmarshal(res2, &tmp4)
	t.Log("tmp3: ", tmp3, "err3: ", err3)
	t.Log("tmp4: ", tmp4, "err4: ", err4)
	t.Log(tmp4.Obj)
}
