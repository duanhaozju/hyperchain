package db_utils

import (
	"hyperchain/core/types"
	"testing"
	"hyperchain/hyperdb"
	"reflect"
	"hyperchain/core/test_util"
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
	logger.Info("test =============> > > TestInitializeChain")
	InitDataBase()
	InitializeChain(hyperdb.DefautNameSpace)

	SetHeightOfChain(hyperdb.DefautNameSpace, 10)
	if GetHeightOfChain(hyperdb.DefautNameSpace) != 10 {
		t.Errorf("SetHeightOfChain and GetHeightOfChain fail")
	}

	SetTxSumOfChain(hyperdb.DefautNameSpace, 10)
	if GetTxSumOfChain(hyperdb.DefautNameSpace) != 10 {
		t.Errorf("SetTxSumOfChain and GetTxSumOfChain fail")
	}

	SetLatestBlockHash(hyperdb.DefautNameSpace, []byte("00000 00000 00000 00000 00000 00000 04"))
	if string(GetLatestBlockHash(hyperdb.DefautNameSpace)) != "00000 00000 00000 00000 00000 00000 04"{
		t.Errorf("SetLatestBlockHash and GetLatestBlockHash fail")
	}

	SetParentBlockHash(hyperdb.DefautNameSpace, []byte("00000 00000 00000 00000 00000 00000 05"))
	if string(GetParentBlockHash(hyperdb.DefautNameSpace)) != "00000 00000 00000 00000 00000 00000 05" {
		t.Errorf("SetParentBlockHash and GetParentBlockHash fail")
	}
	deleteTestData()
}

//TestGetChainCopy tests for GetChainCopy
func TestGetChainCopy(t *testing.T) {
	logger.Info("test =============> > > TestGetChainCopy")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	ch := GetChainCopy(hyperdb.DefautNameSpace)
	if ch.Height != 1 || !reflect.DeepEqual(ch.ParentBlockHash, []byte("parentHash")) {
		t.Errorf("GetChainCopy fail")
	}
	deleteTestData()
}

//TestGetChainUntil tests for GetChainUntil
func TestGetChainUntil(t *testing.T) {
	logger.Info("test =============> > > TestGetChainUntil")
	InitDataBase()
	go WriteChainChan(hyperdb.DefautNameSpace)
	GetChainUntil(hyperdb.DefautNameSpace)
	deleteTestData()
}