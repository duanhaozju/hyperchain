package db_utils

import (
	"hyperchain/core/types"
	"testing"
	"hyperchain/hyperdb"
)

var chain = types.Chain{
	LatestBlockHash : []byte("00000 00000 00000 00000 00000 00000 02"),
	ParentBlockHash : []byte("00000 00000 00000 00000 00000 00000 01"),
	Height         : 2,
	RequiredBlockNum : 3,
	RequireBlockHash : []byte("00000 00000 00000 00000 00000 00000 03"),
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
}

//TestGetChainCopy tests for GetChainCopy
func TestGetChainCopy(t *testing.T) {
	logger.Info("test =============> > > TestGetChainCopy")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
	PersistBlock(db.NewBatch(), &blockCases, true, true)
	UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	ch := GetChainCopy(hyperdb.DefautNameSpace)
	if ch.Height != 1 || ch.CurrentTxSum != 3 {
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