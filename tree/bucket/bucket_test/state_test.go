package bucket_test_test

import (
	"testing"
	"github.com/op/go-logging"
	"github.com/golang/protobuf/proto"
	"hyperchain/tree/bucket/bucket_test"
	"hyperchain/tree/bucket"
	"hyperchain/tree/bucket/testutil"
	"hyperchain/hyperdb"
)
var (
	logger = logging.MustGetLogger("bucket_test")
)
func init(){
}
func TestState_GetHash(t *testing.T){
	state := bucket_test.NewState()
	key_valueMap := make(bucket.K_VMap)
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	state.SetK_VMap(key_valueMap)
	rootHash,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(rootHash))
	}
	// logger.Debug("*******************the cryptoHashForBucket is ",expectedBucketHashForTest([]string{"testStateImpl","chaincodeID1key1", "value1"},))
	expectedHash1 := expectedBucketHashForTest(
		[]string{"testStateImpl","key1", "value1"},
		[]string{"testStateImpl","key2", "value2"},
		[]string{"testStateImpl","key3", "value3"},
		[]string{"testStateImpl","key4", "value4"},
	)
	expectedHash2 := expectedBucketHashForTest(
		[]string{"testStateImpl","key5", "value5"},
		[]string{"testStateImpl","key6", "value6"},
		[]string{"testStateImpl","key7", "value7"},
		[]string{"testStateImpl","key8", "value8"},
	)
	expectedHash := testutil.ComputeCryptoHash(expectedHash1,expectedHash2)

	logger.Debugf("the rootHash is ",rootHash)
	logger.Debugf("the exceptedHash is ",expectedHash)

	testutil.AssertEquals(t, rootHash, expectedHash)
}
func TestState_(t *testing.T){
	db,err := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()
	state := bucket_test.NewState()
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	state.SetK_VMap(key_valueMap)
	hash,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}

	state.AddChangesForPersistence(writeBatch)
	state.CommitStateDelta()
}
func expectedBucketHashContentForTest(data ...[]string) []byte {
	expectedContent := []byte{}
	for _, chaincodeData := range data {
		expectedContent = append(expectedContent, encodeNumberForTest(len(chaincodeData[0]))...)
		expectedContent = append(expectedContent, chaincodeData[0]...)
		expectedContent = append(expectedContent, encodeNumberForTest((len(chaincodeData)-1)/2)...)
		for i := 1; i < len(chaincodeData); i++ {
			expectedContent = append(expectedContent, encodeNumberForTest(len(chaincodeData[i]))...)
			expectedContent = append(expectedContent, chaincodeData[i]...)
		}
	}
	return expectedContent
}

func encodeNumberForTest(i int) []byte {
	return proto.EncodeVarint(uint64(i))
}

func expectedBucketHashForTest(data ...[]string) []byte {
	return testutil.ComputeCryptoHash(expectedBucketHashContentForTest(data...))
}
