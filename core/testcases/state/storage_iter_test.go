package state

import (
	"testing"
	"fmt"
	"hyperchain/core/hyperstate"
	"hyperchain/common"
	"os"
	"path"
	"math/big"
	"hyperchain/hyperdb"
	"hyperchain/core/db_utils"
	"hyperchain/core/vm"
)

var (
	configPath = "./namespaces/global/config/global.yaml"
	globalConfig *common.Config
)


func TestStorageIterator(t *testing.T) {
	opath := switchToExeLoc()
	initLog()
	defer clearup(opath)
	db_utils.InitDBForNamespace(globalConfig, "global")
	hyperdb.StartDatabase(globalConfig, "global")

	db, _ := hyperdb.GetDBDatabaseByNamespace("global")
	stateDb, _ := hyperstate.New(common.Hash{}, db, db, globalConfig, 0, "global")
	stateDb.MarkProcessStart(1)
	stateDb.CreateAccount(common.BytesToAddress([]byte("address001")))
	stateDb.AddBalance(common.BytesToAddress([]byte("address001")), big.NewInt(100))
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key1")), []byte("value1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key2")), []byte("value2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key3")), []byte("value3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key4")), []byte("value4"), 0)
	stateDb.Commit()
	stateDb.Reset()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	stateDb.MarkProcessStart(2)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key1")), []byte("newvalue1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key2")), []byte("newvalue2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key3")), []byte("newvalue3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key5")), []byte("value5"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key6")), []byte("value6"), 0)

	start := common.BytesToHash([]byte("key2"))
	limit := common.BytesToHash([]byte("key5"))

	iter, _ := stateDb.NewIterator(common.BytesToAddress([]byte("address001")), &vm.IterRange{Start: &start, Limit: &limit})
	for iter.Next() {
		fmt.Println("key", common.Bytes2Hex(iter.Key()))
		fmt.Println("value", common.Bytes2Hex(iter.Value()))
	}
	fmt.Println("DONE")
	iter.Release()

	iter, _ = stateDb.NewIterator(common.BytesToAddress([]byte("address001")), &vm.IterRange{Start: nil, Limit: &limit})
	for iter.Next() {
		fmt.Println("key", common.Bytes2Hex(iter.Key()))
		fmt.Println("value", common.Bytes2Hex(iter.Value()))
	}
	fmt.Println("DONE")
	iter.Release()

	iter, _ = stateDb.NewIterator(common.BytesToAddress([]byte("address001")), nil)
	for iter.Next() {
		fmt.Println("key", common.Bytes2Hex(iter.Key()))
		fmt.Println("value", common.Bytes2Hex(iter.Value()))
	}

	fmt.Println("DONE")
	iter.Release()

	iter, _ = stateDb.NewIterator(common.BytesToAddress([]byte("address001")), vm.BytesPrefix(common.BytesToHash([]byte("key0")).Bytes()))
	for iter.Next() {
		fmt.Println("key", common.Bytes2Hex(iter.Key()))
		fmt.Println("value", common.Bytes2Hex(iter.Value()))
	}

}

func switchToExeLoc() string {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/configuration"))
	return owd
}

func switchBack(dir string) {
	os.Chdir(dir)
}

func initLog() {
	globalConfig = common.NewConfig(configPath)
	common.InitHyperLoggerManager(globalConfig)
	globalConfig.Set(common.NAMESPACE, "global")
	common.InitHyperLogger(globalConfig)
}

func clearup(dir string) {
	os.RemoveAll("./namespaces/global/data")
	switchBack(dir)
}