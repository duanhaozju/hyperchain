package db_utils

import (
	"hyperchain/core/types"
	"testing"
	"hyperchain/hyperdb"
	"reflect"
	"hyperchain/core/test_util"
	"hyperchain/common"
)


var chain = types.Chain{
	LatestBlockHash : []byte("00000000000000000000000000000002"),
	ParentBlockHash : []byte("00000000000000000000000000000001"),
	Height         : 2,
	RequiredBlockNum : 3,
	RequireBlockHash : []byte("00000000000000000000000000000003"),
	RecoveryNum      : 2,
	CurrentTxSum     : 100,
}

var mC = memChain{
	data    :chain,
	cpChan  : make(chan types.Chain),
	txDelta : 10,
}

var mCs = memChains {
	chains : make(map[string]*memChain),
}

//TestInitializeChain tests for InitializeChain
func TestInitializeChain(t *testing.T) {
	t.Log("test =============> > > TestInitializeChain")
	InitDataBase()
	InitializeChain(common.DEFAULT_NAMESPACE)

	SetHeightOfChain(common.DEFAULT_NAMESPACE, 10)
	if GetHeightOfChain(common.DEFAULT_NAMESPACE) != 10 {
		t.Errorf("SetHeightOfChain and GetHeightOfChain fail")
	}

	SetTxSumOfChain(common.DEFAULT_NAMESPACE, 10)
	if GetTxSumOfChain(common.DEFAULT_NAMESPACE) != 10 {
		t.Errorf("SetTxSumOfChain and GetTxSumOfChain fail")
	}

	SetLatestBlockHash(common.DEFAULT_NAMESPACE, []byte("00000 00000 00000 00000 00000 00000 04"))
	if string(GetLatestBlockHash(common.DEFAULT_NAMESPACE)) != "00000 00000 00000 00000 00000 00000 04"{
		t.Errorf("SetLatestBlockHash and GetLatestBlockHash fail")
	}

	SetParentBlockHash(common.DEFAULT_NAMESPACE, []byte("00000 00000 00000 00000 00000 00000 05"))
	if string(GetParentBlockHash(common.DEFAULT_NAMESPACE)) != "00000 00000 00000 00000 00000 00000 05" {
		t.Errorf("SetParentBlockHash and GetParentBlockHash fail")
	}
	deleteTestData()
}

//TestGetChainCopy tests for GetChainCopy
func TestGetChainCopy(t *testing.T) {
	t.Log("test =============> > > TestGetChainCopy")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	UpdateChainByBlcokNum(common.DEFAULT_NAMESPACE, db.NewBatch(), 1, true, true)
	ch := GetChainCopy(common.DEFAULT_NAMESPACE)
	if ch.Height != 1 || !reflect.DeepEqual(ch.ParentBlockHash, []byte("parentHash")) {
		t.Errorf("GetChainCopy fail")
	}
	deleteTestData()
}

//TestGetChainUntil tests for GetChainUntil
func TestGetChainUntil(t *testing.T) {
	t.Log("test =============> > > TestGetChainUntil")
	InitDataBase()
	go WriteChainChan(common.DEFAULT_NAMESPACE)
	GetChainUntil(common.DEFAULT_NAMESPACE)
	deleteTestData()
}