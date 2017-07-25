package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"hyperchain/core/test_util"
	"gopkg.in/check.v1"
	"hyperchain/core/types"
	"reflect"
	"math/rand"
	"testing/quick"
	"time"
	"strconv"
	"hyperchain/common"
)


func Test(t *testing.T) {
	check.TestingT(t)
}

type blockSuite struct {
	block types.Block
}

var _ = check.Suite(&blockSuite{})

func (s *blockSuite) SetUpTest(c *check.C) {
	InitDataBase()
}

func (s *blockSuite) TearDownTest(c *check.C) {
	deleteTestData()
}

func (s *blockSuite) TestBlockSuite1(c *check.C) {
	c.Assert(GetVersionOfBlock(), check.Equals, BlockVersion)
}

func (s *blockSuite) TestBlockSuite2(c *check.C) {
	c.Assert(GetParentHashOfBlock(), check.Equals, "parentHash")
}

func (s *blockSuite) TestBlockSuite3(c *check.C) {
	c.Assert(GetBlockHashOfBlock(), check.Equals, "blockHash")
}

func (s *blockSuite) TestBlockSuite4(c *check.C) {
	c.Assert(GetTransactionsOfBlock(), check.DeepEquals, test_util.TransactionCases)
}

func (s *blockSuite) TestBlockSuite5(c *check.C) {
	c.Assert(GetTimestampOfBlock(), check.Equals, int64(1489387222))
}

func (s *blockSuite) TestBlockSuite6(c *check.C) {
	c.Assert(GetMerkleRootOfBlock(), check.Equals, "merkleRoot")
}

func (s *blockSuite) TestBlockSuite7(c *check.C) {
	c.Assert(GetTxRootOfBlock(), check.Equals, "txRoot")
}

func (s *blockSuite) TestBlockSuite8(c *check.C) {
	c.Assert(GetReceiptRootOfBlock(), check.Equals, "receiptRoot")
}

func (s *blockSuite) TestBlockSuite9(c *check.C) {
	c.Assert(GetNumberOfBlock(), check.Equals, uint64(1))
}

func (s *blockSuite) TestBlockSuite10(c *check.C) {
	c.Assert(GetWriteTimeOfBlock(), check.Equals, int64(1489387223))
}

func (s *blockSuite) TestBlockSuite11(c *check.C) {
	c.Assert(GetCommitTimeOfBlock(), check.Equals, int64(1489387224))
}

func (s *blockSuite) TestBlockSuite12(c *check.C) {
	c.Assert(GetEvmTimeOfBlock(), check.Equals, int64(1489387225))
}

func GetVersionOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return string(block.Version)
}

func GetParentHashOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return string(block.ParentHash)
}

func GetBlockHashOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	return string(block.BlockHash)
}

func GetTransactionsOfBlock() []*types.Transaction {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.Transactions
}

func GetTimestampOfBlock() int64 {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.Timestamp

}

func GetMerkleRootOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return string(block.MerkleRoot)
}

func GetTxRootOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return string(block.TxRoot)
}

func GetReceiptRootOfBlock() string {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return string(block.ReceiptRoot)
}

func GetNumberOfBlock() uint64 {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.Number
}

func GetWriteTimeOfBlock() int64 {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.WriteTime
}

func GetCommitTimeOfBlock() int64 {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.CommitTime
}

func GetEvmTimeOfBlock() int64 {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	block, _ := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	return block.EvmTime
}



// using quick for unit test
func (blockSuite) Generate(r *rand.Rand, size int) reflect.Value {
	str := "blockHash" + strconv.Itoa(time.Now().Local().Nanosecond())
	blockCases := types.Block{
		Version:      []byte(BlockVersion),
		ParentHash:   []byte("parentHash"),
		BlockHash:    []byte(str),
		Transactions: test_util.TransactionCases,
		Timestamp:    time.Now().Local().UnixNano(),
		MerkleRoot:   []byte("merkleRoot"),
		TxRoot:       []byte("txRoot"),
		ReceiptRoot:  []byte("receiptRoot"),
		Number:       uint64(r.Int()),
		WriteTime:    time.Now().Local().UnixNano(),
		CommitTime:   time.Now().Local().UnixNano(),
		EvmTime:      time.Now().Local().UnixNano(),
	}
	block := blockSuite{block:blockCases}
	return reflect.ValueOf(block)
}

func runBlockTest(block blockSuite) bool {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err,_ := PersistBlock(db.NewBatch(), &block.block, true, true)
	if err != nil {
		return false
	}
	actual, err := GetBlock(common.DEFAULT_NAMESPACE, block.block.BlockHash)
	if !reflect.DeepEqual(actual, &block.block) {
		return false
	}
	deleteTestData()
	return true
}

func TestQuick(t *testing.T) {
	if err := quick.Check(runBlockTest, nil); err != nil {
		t.Error("Test Quick fail", err)
	}
}