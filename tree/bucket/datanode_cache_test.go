package bucket

import (
	"fmt"
	"github.com/spf13/viper"
	"hyperchain/tree/bucket/testutil"
	"testing"
)

var (
	configs map[string]interface{}
)

func init() {
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

	dataNodeCache := newDataNodeCache("-bucket-state", 10)
	key1 := string("key1")
	dataKey1 := newDataKey(dataNodeCache.TreePrefix, key1)
	value1 := []byte("value1")
	dataNode1 := newDataNode(dataKey1, value1)

	//dataNodeCache.Remove(dataNode2)

	dataNodes1, err := dataNodeCache.FetchDataNodesFromCache(*dataNode1.dataKey.bucketKey)
	if err != nil {
		log.Errorf("******error is ", err)
	} else {
		log.Critical("******len is", len(dataNodes1))
		for _, v := range dataNodes1 {
			log.Critical("******dataKey is", v.dataKey)
			log.Critical("******value is", string(v.value))
		}
	}

}
