package bucket

import (
	"testing"
	"hyperchain/tree/bucket/testutil"
	"fmt"
	"github.com/spf13/viper"
)
var(
configs map[string]interface{}
)
func init(){
	configs = viper.GetStringMap("ledger.state.dataStructure.configs")
	stateImpl := NewBucketTree("-bucket-state")
	err := stateImpl.Initialize(configs)
	if err != nil {
		panic(fmt.Errorf("Error during initialization of state implementation: %s", err))
	}
}
func TestDataNodeCache_Put(t *testing.T) {
	testDBWrapper := testutil.NewTestDBWrapper()
	testDBWrapper.CleanDB(t)

	dataNodeCache := newDataNodeCache("-bucket-state",10)
	key1 := string("key1")
	dataKey1 := newDataKey(dataNodeCache.TreePrefix,key1)
	value1 := []byte("value1")
	dataNode1 := newDataNode(dataKey1,value1)

	key2 := string("key1")
	dataKey2 := newDataKey(dataNodeCache.TreePrefix,key2)
	value2 := []byte("value1")
	dataNode2 := newDataNode(dataKey2,value2)


	dataNodeCache.Put(dataNode1)
	dataNodeCache.Put(dataNode2)
	//dataNodeCache.Remove(dataNode2)

	dataNodes1,err := dataNodeCache.FetchDataNodesFromCache(*dataNode1.dataKey.bucketKey)
	if err!= nil{
		logger.Errorf("******error is ",err)
	}else {
		logger.Critical("******len is",len(dataNodes1))
		for _,v := range dataNodes1{
			logger.Critical("******dataKey is",v.dataKey)
			logger.Critical("******value is",string(v.value))
		}
	}

}