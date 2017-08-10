package hyperstate

import (
	"testing"
	"hyperchain/hyperdb/mdb"
	"hyperchain/common"
	"fmt"
	// "hyperchain/core/vm"
	tutil "hyperchain/core/test_util"
	"hyperchain/core/vm"
)

func TestStorageIterator(t *testing.T) {
	configPath := "../../configuration/namespaces/global/config/namespace.toml"
	common.InitHyperLoggerManager(tutil.InitConfig(configPath))
	common.InitRawHyperLogger(common.DEFAULT_NAMESPACE)
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	conf := common.NewRawConfig()

	conf.Set(common.LOG_BASE_LOG_LEVEL, "NOTICE")
	conf.Set(StateBucketSize, 1000003)
	conf.Set(StateBucketLevelGroup, 5)
	conf.Set(StateBucketCacheSize, 100000)
	conf.Set(StateDataNodeCacheSize, 100000)

	conf.Set(StateObjectBucketSize, 1000003)
	conf.Set(StateObjectBucketLevelGroup, 5)
	conf.Set(StateObjectBucketCacheSize, 100000)
	conf.Set(StateObjectDataNodeCacheSize, 100000)

	conf.Set(GlobalDataNodeCacheLength, 100)
	conf.Set(GlobalDataNodeCacheSize, 100)

	stateDb, _ := New(common.Hash{}, db, db, conf, 0, common.DEFAULT_NAMESPACE)
	stateDb.MarkProcessStart(1)
	stateDb.CreateAccount(common.BytesToAddress([]byte("address001")))
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key1")), []byte("value1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key2")), []byte("value2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key3")), []byte("value3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key4")), []byte("value4"), 0)
	stateDb.Commit()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	stateDb.MarkProcessStart(2)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key1")), []byte("newvalue1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key2")), []byte("newvalue2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key3")), []byte("newvalue3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key5")), []byte("newvalue5"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key6")), []byte("newvalue6"), 0)

	//start := common.BytesToRightPaddingHash([]byte("key0"))
	//limit := common.BytesToRightPaddingHash([]byte("key10"))
	iter, _ := stateDb.NewIterator(common.BytesToAddress([]byte("address001")), nil)
	for iter.Next() {
		fmt.Println(common.Bytes2Hex(iter.Key()))
		fmt.Println(common.Bytes2Hex(iter.Value()))
	}
}

func TestStorageIteratorWithPrefix(t *testing.T) {
	configPath := "../../configuration/namespaces/global/config/namespace.toml"
	common.InitHyperLoggerManager(tutil.InitConfig(configPath))
	common.InitRawHyperLogger(common.DEFAULT_NAMESPACE)
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	conf := common.NewRawConfig()

	conf.Set(common.LOG_BASE_LOG_LEVEL, "NOTICE")
	conf.Set(StateBucketSize, 1000003)
	conf.Set(StateBucketLevelGroup, 5)
	conf.Set(StateBucketCacheSize, 100000)
	conf.Set(StateDataNodeCacheSize, 100000)

	conf.Set(StateObjectBucketSize, 1000003)
	conf.Set(StateObjectBucketLevelGroup, 5)
	conf.Set(StateObjectBucketCacheSize, 100000)
	conf.Set(StateObjectDataNodeCacheSize, 100000)

	conf.Set(GlobalDataNodeCacheLength, 100)
	conf.Set(GlobalDataNodeCacheSize, 100)

	stateDb, _ := New(common.Hash{}, db, db, conf, 0, common.DEFAULT_NAMESPACE)
	stateDb.MarkProcessStart(1)
	stateDb.CreateAccount(common.BytesToAddress([]byte("address001")))
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key1")), []byte("value1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key2")), []byte("value2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key3")), []byte("value3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key4")), []byte("value4"), 0)
	stateDb.Commit()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	stateDb.MarkProcessStart(2)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key1")), []byte("newvalue1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key2")), []byte("newvalue2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key3")), []byte("newvalue3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key5")), []byte("newvalue5"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToRightPaddingHash([]byte("key6")), []byte("newvalue6"), 0)

	//start := common.BytesToRightPaddingHash([]byte("key0"))
	//limit := common.BytesToRightPaddingHash([]byte("key10"))
	iter, _ := stateDb.NewIterator(common.BytesToAddress([]byte("address001")), vm.BytesPrefix([]byte("key")))
	for iter.Next() {
		fmt.Println(common.Bytes2Hex(iter.Key()))
		fmt.Println(common.Bytes2Hex(iter.Value()))
	}
}
