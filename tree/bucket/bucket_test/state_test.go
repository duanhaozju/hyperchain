package bucket_test_test

import (
	"testing"
	"github.com/op/go-logging"
	"github.com/golang/protobuf/proto"
	"hyperchain/tree/bucket/bucket_test"
	"hyperchain/tree/bucket"
	"hyperchain/tree/bucket/testutil"
	"hyperchain/hyperdb"
	"math/big"
	"hyperchain/common"
)
var (
	logger = logging.MustGetLogger("bucket_test")
)
func init(){
}
func TestState_GetHash(t *testing.T){
	state := bucket_test.NewState("TestState_GetHash")
	key_valueMap := make(bucket.K_VMap)
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	state.SetK_VMap(key_valueMap,big.NewInt(2))
	rootHash,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------state.GetHash() is ",(rootHash))
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
func TestState_1_simple(t *testing.T){
	db,err := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()
	state := bucket_test.NewState("TestState_1_simple")
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	key_valueMap["key9"] = []byte("value9")
	key_valueMap["key10"] = []byte("value10")
	state.SetK_VMap(key_valueMap,big.NewInt(2))
	hash,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}
	state.CommitStateDelta(writeBatch,big.NewInt(2))
}

func TestState_2_without_persist(t *testing.T){
	db,err := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()
	state := bucket_test.NewState("TestState_2_without_persist")
	// block 2
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(2))
	hash,err := state.Bucket_tree.ComputeCryptoHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(2))
	//writeBatch.Write()
	// block 3
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	key_valueMap["key9"] = []byte("value9")
	key_valueMap["key10"] = []byte("value10")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(3))
	hash,err = state.Bucket_tree.ComputeCryptoHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(3))
	//writeBatch.Write()

	// block 4
	key_valueMap = bucket.K_VMap{}
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(4))
	hash,err = state.Bucket_tree.ComputeCryptoHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(4))
	//writeBatch.Write()


	// block 5
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key11"] = []byte("value11")
	key_valueMap["key12"] = []byte("value12")
	key_valueMap["key13"] = []byte("value13")
	key_valueMap["key14"] = []byte("value14")
	key_valueMap["key15"] = []byte("value15")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(5))
	hash,err = state.Bucket_tree.ComputeCryptoHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(5))

}

func TestState_3_persist(t *testing.T){
	testDBWrapper := testutil.NewTestDBWrapper()
	testDBWrapper.CleanDB(t)

	db,err := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()

	state := bucket_test.NewState("TestState_3_persist")
	hash1,err := state.GetHash()
	testutil.AssertEquals(t,[]byte(nil),hash1)

	// block 2
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	key_valueMap["key5"] = []byte("value5")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(2))
	hash2,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash2))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(2))
	writeBatch.Write()
	testutil.AssertNotEquals(t,hash1,hash2)
	// block 3
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	key_valueMap["key9"] = []byte("value9")
	key_valueMap["key10"] = []byte("value10")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(3))
	hash3,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash3))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(3))
	writeBatch.Write()
	testutil.AssertNotEquals(t,hash2,hash3)

	// block 4
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key6"] = []byte("value6")
	key_valueMap["key7"] = []byte("value7")
	key_valueMap["key8"] = []byte("value8")
	key_valueMap["key9"] = []byte("value9")
	key_valueMap["key10"] = []byte("value10")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(4))
	hash4,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash4))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(4))
	writeBatch.Write()
	testutil.AssertEquals(t,hash3,hash4)


	// block 5
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key10"] = []byte("value10-2")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(5))
	hash5,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash5))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(5))
	writeBatch.Write()

	// block 6
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key10"] = []byte("value10-2")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(6))
	hash6,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash6))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(6))
	writeBatch.Write()
	testutil.AssertEquals(t,hash5,hash6)

	// block 7
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key10"] = []byte("value10-2")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(7))
	hash7,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the state.GetHash() is ",(hash7))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch,big.NewInt(7))
	writeBatch.Write()
	testutil.AssertEquals(t,hash5,hash6)



}

func TestRevertToTargetBlock(t *testing.T) {
	testDBWrapper := testutil.NewTestDBWrapper()
	testDBWrapper.CleanDB(t)

	db,err := hyperdb.GetLDBDatabase()
	writeBatch0 := db.NewBatch()

	state := bucket_test.NewState("-bucket-state")
	// block 0
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key1"] = []byte("value1")
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(0))
	hash0,err := state.GetHash()
	//treeHash1,err := state.Bucket_tree.GetTreeHash(big.NewInt(1))
	logger.Debugf("--------------the state.GetHash() is ",(hash0))

	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash0 is ",common.Bytes2Hex(hash0))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch0,big.NewInt(0))
	writeBatch0.Write()

	// block 1
	writeBatch1 := db.NewBatch()
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key2"] = []byte("value2-2")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(1))
	hash1,err := state.GetHash()
	//treeHash1,err := state.Bucket_tree.GetTreeHash(big.NewInt(1))
	logger.Debugf("--------------the state.GetHash() is ",(hash1))

	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash1 is ",common.Bytes2Hex(hash1))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch1,big.NewInt(1))
	writeBatch1.Write()


	// block 2
	writeBatch2 := db.NewBatch()
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key2"] = []byte("value2-4")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(2))
	hash2,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		//logger.Debugf("--------------the hash2 is ",common.Bytes2Hex(hash2))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch2,big.NewInt(2))
	//writeBatch.Write()

	// block 3
	writeBatch3 := db.NewBatch()
	key_valueMap = bucket.NewKVMap()
	key_valueMap["key21"] = []byte("value2-3")

	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(3))
	//hash3,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		//logger.Debugf("--------------the hash3 is ",common.Bytes2Hex(hash3))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch3,big.NewInt(3))

	// block 4
	writeBatch4 := db.NewBatch()
	key_valueMap = bucket.NewKVMap()
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(4))
	hash4,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash4 is ",common.Bytes2Hex(hash4))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch4,big.NewInt(4))
	writeBatch2.Write()
	writeBatch3.Write()
	writeBatch4.Write()

	// block 6
	writeBatch6 := db.NewBatch()
	key_valueMap = bucket.NewKVMap()
	key_valueMap["key3"] = []byte("value3-5")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(6))
	hash6,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash6 is ",common.Bytes2Hex(hash6))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch6,big.NewInt(6))
	writeBatch6.Write()


	state = bucket_test.NewState("-bucket-state")
	fromBlock := big.NewInt(6)
	toBlock := big.NewInt(1)
	state.Bucket_tree.RevertToTargetBlock(fromBlock,toBlock)
	hash_revert1,err := state.GetHash()
	logger.Debugf("after revert from %d",fromBlock," to %d",toBlock,"--------------the state.GetHash() is ",common.Bytes2Hex(hash_revert1))
	testutil.AssertEquals(t,hash1,hash_revert1)

	// block 2
	writeBatch2 = db.NewBatch()
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key2"] = []byte("value2-4")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(2))
	newHash2,err := state.GetHash()
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		//logger.Debugf("--------------the hash2 is ",common.Bytes2Hex(hash2))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch2,big.NewInt(2))
	testutil.AssertEquals(t,hash2,newHash2)

	//
	//state = bucket_test.NewState()
	//fromBlock = big.NewInt(4)
	//toBlock = big.NewInt(0)
	//state.Bucket_tree.RevertToTargetBlock(fromBlock,toBlock)
	//hash_revert2,err := state.GetHash()
	//logger.Debugf("after revert from %d",fromBlock," to %d",toBlock,"--------------the state.GetHash() is ",common.Bytes2Hex(hash_revert2))
	//testutil.AssertEquals(t,hash2,hash_revert2)
}

func TestRevertToTargetBlock_2(t *testing.T) {
	testDBWrapper := testutil.NewTestDBWrapper()
	testDBWrapper.CleanDB(t)

	db,err := hyperdb.GetLDBDatabase()
	writeBatch0 := db.NewBatch()

	state := bucket_test.NewState("TestState")
	state2 := bucket_test.NewState("TestState222")
	// block 0
	key_valueMap := bucket.K_VMap{}
	key_valueMap["key2"] = []byte("value2")
	key_valueMap["key1"] = []byte("value1")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(0))
	hash0,err := state.GetHash()
	logger.Debugf("--------------the state.GetHash() is ",(hash0))

	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash0 is ",common.Bytes2Hex(hash0))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch0,big.NewInt(0))
	writeBatch0.Write()

	key_valueMap = bucket.K_VMap{}
	key_valueMap["key2-2"] = []byte("value2-2")
	key_valueMap["key1-2"] = []byte("value1-2")
	state2.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(0))
	hash0_2,err := state.GetHash()
	//treeHash1,err := state.Bucket_tree.GetTreeHash(big.NewInt(1))
	logger.Debugf("--------------the state.GetHash() is ",(hash0_2))
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash0 is ",common.Bytes2Hex(hash0_2))
	}
	state2.Bucket_tree.AddChangesForPersistence(writeBatch0,big.NewInt(0))
	writeBatch0.Write()


	// block 1
	key_valueMap = bucket.K_VMap{}
	key_valueMap["key3"] = []byte("value3")
	key_valueMap["key4"] = []byte("value4")
	state.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(1))
	hash1,err := state.GetHash()
	logger.Debugf("--------------the state.GetHash() is ",(hash0))

	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash0 is ",common.Bytes2Hex(hash1))
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch0,big.NewInt(1))
	writeBatch0.Write()

	key_valueMap = bucket.K_VMap{}
	key_valueMap["key3-2"] = []byte("value3-2")
	key_valueMap["key4-2"] = []byte("value4-2")
	state2.Bucket_tree.PrepareWorkingSet(key_valueMap,big.NewInt(1))
	hash1_2,err := state.GetHash()
	//treeHash1,err := state.Bucket_tree.GetTreeHash(big.NewInt(1))
	logger.Debugf("--------------the state.GetHash() is ",(hash0_2))
	if err != nil{
		logger.Debugf("--------------GetHash error")
	}else {
		logger.Debugf("--------------the hash0 is ",common.Bytes2Hex(hash1_2))
	}
	state2.Bucket_tree.AddChangesForPersistence(writeBatch0,big.NewInt(1))
	writeBatch0.Write()


	state = bucket_test.NewState("TestState222")
	fromBlock := big.NewInt(2)
	toBlock := big.NewInt(1)
	state.Bucket_tree.RevertToTargetBlock(fromBlock,toBlock)
	hash_revert1,err := state.GetHash()
	logger.Debugf("after revert from %d",fromBlock," to %d",toBlock,"--------------the state.GetHash() is ",common.Bytes2Hex(hash_revert1))
	testutil.AssertEquals(t,hash1_2,hash_revert1)
	//
	//state = bucket_test.NewState()
	//fromBlock = big.NewInt(4)
	//toBlock = big.NewInt(0)
	//state.Bucket_tree.RevertToTargetBlock(fromBlock,toBlock)
	//hash_revert2,err := state.GetHash()
	//logger.Debugf("after revert from %d",fromBlock," to %d",toBlock,"--------------the state.GetHash() is ",common.Bytes2Hex(hash_revert2))
	//testutil.AssertEquals(t,hash2,hash_revert2)
}

func featchDataNodeFromDBTest(){
	//db,_ := hyperdb.GetLDBDatabase()
	//db.Get()
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
