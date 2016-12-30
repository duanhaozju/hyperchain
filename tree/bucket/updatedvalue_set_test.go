package bucket

import (
	"testing"
	"github.com/golang/protobuf/proto"
	"math/big"
	"hyperchain/tree/bucket/testutil"
)

func TestUpdatedValueSet_Marshal(t *testing.T) {
	buffer := proto.NewBuffer([]byte{})

	updatevalueSet := newUpdatedValueSet(big.NewInt(1))
	newUpdatevalueSet := newUpdatedValueSet(big.NewInt(1))
	updatevalueSet.Set("key1",[]byte("previous value1"),[]byte("current value1"))
	updatevalueSet.Set("key2",[]byte("previous value2"),[]byte("current value2"))
	updatevalueSet.Set("key3",[]byte("previous value3"),[]byte("current value3"))
	updatevalueSet.Set("key4",[]byte("previous value4"),[]byte("current value4"))
	updatevalueSet.Marshal(buffer)
	updatevalueSet.Print("TestUpdatedValueSet_Marshal")
	newUpdatevalueSet.UnMarshal(buffer)
	newUpdatevalueSet.Print("TestUpdatedValueSet_Marshal")
	testutil.AssertEquals(t,newUpdatevalueSet,updatevalueSet)
}