package core

import (
	"testing"
	"log"
	"hyperchain/common"
	"hyperchain/core/types"
)


// TestGetBalanceIns tests for GetBalanceIns
func TestGetBalanceIns(t *testing.T) {
	log.Println("test =============> > > TestGetBalanceIns")
	quit := make(chan int)
	GetBalanceIns()
	for i := 0; i < 10; i++ {
		go func() {
			_, err := GetBalanceIns()
			if err != nil {
				log.Fatal(err)
			}
			quit <- 0
		}()
	}
	for i := 0; i < 10; i++ {
		<- quit
	}
}

var balanceCases = BalanceMap{
	common.StringToAddress("zhangsan"): []byte("1000"),
	common.StringToAddress("lisi") : []byte("2000"),
	common.StringToAddress("wangwu") : []byte("3000"),
}

// TestBalance_PutCacheBalance tests for PutCacheBalance
func TestBalance_PutCacheBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_PutCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		b.PutCacheBalance(key, value)
	}
}

// TestBalance_GetCacheBalance tests for GetCacheBalance
func TestBalance_GetCacheBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_GetCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		data := b.GetCacheBalance(key)
		if string(value) != string(data) {
			t.Errorf("%s not equal %s, TestBalance_GetCacheBalance fail", string(value), string(data))
		}
	}
}

// TestBalance_GetAllCacheBalance tests for GetAllCacheBalance
func TestBalance_GetAllCacheBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_GetAllCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	cacheBs := b.GetAllCacheBalance()
	for key, value := range balanceCases {
		data := cacheBs[key]
		if string(value) != string(data) {
			t.Errorf("%s not equal %s, TestBalance_GetAllCacheBalance fail", string(value), string(data))
		}
	}
}

func TestBalance_DeleteCacheBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_DeleteCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, _ := range balanceCases {
		b.DeleteCacheBalance(key)
		data := b.GetCacheBalance(key)
		if len(data) != 0 {
			t.Errorf("the len of %v is not equal 0, TestBalance_DeleteCacheBalance fail", data)
		}
	}
}

// TestBalance_PutDBBalance tests for PutDBBalance
func TestBalance_PutDBBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_PutDBBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		b.PutDBBalance(key, value)
	}
}

// TestBalance_GetDBBalance tests for GetDBBalance
func TestBalance_GetDBBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_GetDBBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		data := b.GetDBBalance(key)
		if string(value) != string(data) {
			t.Errorf("%s not equal %s, TestBalance_GetDBBalance fail", string(value), string(data))
		}
	}
}

// TestBalance_GetAllDBBalance tests for GetAllDBBalance
func TestBalance_GetAllDBBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_GetAllDBBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	DBBs := b.GetAllDBBalance()
	for key, value := range balanceCases {
		data := DBBs[key]
		if string(value) != string(data) {
			t.Errorf("%s not equal %s, TestBalance_GetAllDBBalance fail", string(value), string(data))
		}
	}
}

func TestBalance_DeleteDBBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_DeleteDBBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, _ := range balanceCases {
		b.DeleteDBBalance(key)
		data := b.GetDBBalance(key)
		if len(data) != 0 {
			t.Errorf("the len of %v is not equal 0, TestBalance_DeleteDBBalance fail", data)
		}
	}
}

var transCases = []*types.Transaction{
	&types.Transaction{
		From: []byte("zhangsan"),
		To: []byte("wangwu"),
		Value: []byte("100"),
	},
	&types.Transaction{
		From: []byte("zhangsan"),
		To: []byte("lisi"),
		Value: []byte("100"),
	},
	&types.Transaction{
		From: []byte("lisi"),
		To: []byte("wangwu"),
		Value: []byte("700"),
	},
}

func TestBalance_UpdateCacheBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_UpdateCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		b.PutCacheBalance(key, value)
	}
	for _, trans := range transCases {
		b.UpdateCacheBalance(trans)
	}
	zhangsan := b.GetCacheBalance(common.StringToAddress("zhangsan"))
	lisi := b.GetCacheBalance(common.StringToAddress("lisi"))
	wangwu := b.GetCacheBalance(common.StringToAddress("wangwu"))
	if string(zhangsan) != "800" || string(lisi) != "1400" || string(wangwu) != "3800"{
		t.Errorf("TestBalance_UpdateCacheBalance fail")
	}
	for key, _ := range balanceCases {
		b.DeleteCacheBalance(key)
	}
}

var blockCase = &types.Block{
	ParentHash: []byte("parenthash"),
	BlockHash: []byte("blockhash"),
	Transactions: transCases,
}

func TestBalance_UpdateDBBalance(t *testing.T) {
	log.Println("test =============> > > TestBalance_UpdateCacheBalance")
	b, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range balanceCases {
		b.PutDBBalance(key, value)
	}
	b.UpdateDBBalance(blockCase)

	zhangsan1 := b.GetCacheBalance(common.StringToAddress("zhangsan"))
	lisi1 := b.GetCacheBalance(common.StringToAddress("lisi"))
	wangwu1 := b.GetCacheBalance(common.StringToAddress("wangwu"))

	zhangsan2 := b.GetCacheBalance(common.StringToAddress("zhangsan"))
	lisi2 := b.GetCacheBalance(common.StringToAddress("lisi"))
	wangwu2 := b.GetCacheBalance(common.StringToAddress("wangwu"))

	if string(zhangsan1) != "800" || string(lisi1) != "1400" || string(wangwu1) != "3800"{
		t.Errorf("TestBalance_UpdateDBBalance fail")
	}
	if string(zhangsan2) != "800" || string(lisi2) != "1400" || string(wangwu2) != "3800"{
		t.Errorf("TestBalance_UpdateDBBalance fail")
	}
	for key, _ := range balanceCases {
		b.DeleteCacheBalance(key)
	}
	for key, _ := range balanceCases {
		b.DeleteDBBalance(key)
	}
}