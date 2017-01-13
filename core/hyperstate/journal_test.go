package hyperstate

import (
	"fmt"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"math/big"
	"testing"
)

type JournalSuite struct {
}

func Test(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&JournalSuite{})

func (s *JournalSuite) TestMarshal(c *checker.C) {
	var jourList []JournalEntry
	addr1 := common.BytesToAddress([]byte("1"))
	addr2 := common.BytesToAddress([]byte("2"))
	addr3 := common.BytesToAddress([]byte("3"))
	addr4 := common.BytesToAddress([]byte("4"))
	addr5 := common.BytesToAddress([]byte("5"))
	addr6 := common.BytesToAddress([]byte("6"))
	addr9 := common.BytesToAddress([]byte("9"))
	addr10 := common.BytesToAddress([]byte("10"))
	jourList = append(jourList, &CreateObjectChange{
		Account: &addr1,
	})
	jourList = append(jourList, &SuicideChange{
		Account:     &addr2,
		Prev:        true,
		Prevbalance: big.NewInt(2),
	})
	jourList = append(jourList, &BalanceChange{
		Account: &addr3,
		Prev:    big.NewInt(3),
	})
	jourList = append(jourList, &NonceChange{
		Account: &addr4,
		Prev:    uint64(4),
	})
	jourList = append(jourList, &StorageChange{
		Account:  &addr5,
		Key:      common.BytesToHash([]byte("key5")),
		Prevalue: common.BytesToHash([]byte("value5")),
	})
	jourList = append(jourList, &CodeChange{
		Account:  &addr6,
		Prevcode: []byte("code6"),
		Prevhash: []byte("hash6"),
	})
	jourList = append(jourList, &RefundChange{
		Prev: big.NewInt(7),
	})
	jourList = append(jourList, &AddLogChange{
		Txhash: common.BytesToHash([]byte("txhash8")),
	})
	jourList = append(jourList, &TouchChange{
		Account: &addr9,
		Prev:    true,
	})
	so := &StateObject{
		address: addr10,
		data: Account{
			Nonce:    10,
			Balance:  big.NewInt(10),
			Root:     common.BytesToHash([]byte("hash10")),
			CodeHash: []byte("codehash10"),
		},
	}
	jourList = append(jourList, &ResetObjectChange{
		Prev: so,
	})

	j := &Journal{JournalList: jourList}
	fmt.Println(j)
	res, err := j.Marshal()
	if err != nil {
		c.Error("journal marshal err")
	}
	jo, err := UnmarshalJournal(res)
	fmt.Println("err: ", err, "jo: ", jo)
}
