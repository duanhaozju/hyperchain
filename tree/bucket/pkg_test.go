package bucket

import (
	"testing"
	"os"
	"BucketTree/bucket/testutil"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/big"
)
type stateImplTestWrapper struct {
	configMap map[string]interface{}
	stateImpl *BucketTree
	t         testing.TB
}

var testDBWrapper = testutil.NewTestDBWrapper()
var testParams []string

func TestMain(m *testing.M) {
	testParams = testutil.ParseTestParams()
	testutil.SetupTestConfig()
	os.Exit(m.Run())
}


func newStateImplTestWrapper(t testing.TB,accountID string) *stateImplTestWrapper {
	var configMap map[string]interface{}
	stateImpl := NewBucketTree(accountID)
	err := stateImpl.Initialize(configMap)
	testutil.AssertNoError(t, err, "Error while constrcuting stateImpl")
	return &stateImplTestWrapper{configMap, stateImpl, t}
}

// testHasher is a hash function for testing.
// It returns the hash for a key from pre-populated map
type testHasher struct {
	testHashFunctionInput map[string]uint32
}

func newTestHasher() *testHasher {
	return &testHasher{make(map[string]uint32)}
}

func (testHasher *testHasher) populate(chaincodeID string, key string, hash uint32) {
	testHasher.testHashFunctionInput[string(ConstructCompositeKey(chaincodeID, key))] = hash
}

func (testHasher *testHasher) getHashFunction() hashFunc {
	return func(data []byte) uint32 {
		key := string(data)
		value, ok := testHasher.testHashFunctionInput[key]
		if !ok {
			panic(fmt.Sprintf("A test should add entry before looking up. Entry looked up = [%s]", key))
		}
		return value
	}
}

func (testWrapper *stateImplTestWrapper) prepareWorkingSetAndComputeCryptoHash(key_valueMap K_VMap) []byte {
	testWrapper.prepareWorkingSet(key_valueMap)
	return testWrapper.computeCryptoHash()
}


func (testWrapper *stateImplTestWrapper) prepareWorkingSet(key_valueMap K_VMap) {
	err := testWrapper.stateImpl.PrepareWorkingSet(key_valueMap,big.NewInt(1))
	testutil.AssertNoError(testWrapper.t, err, "Error while PrepareWorkingSet")
}

func (testWrapper *stateImplTestWrapper) computeCryptoHash() []byte {
	cryptoHash, err := testWrapper.stateImpl.ComputeCryptoHash()
	testutil.AssertNoError(testWrapper.t, err, "Error while computing crypto hash")
	return cryptoHash
}


func createFreshDBAndInitTestStateImplWithCustomHasher(t testing.TB, numBuckets int, maxGroupingAtEachLevel int,testAccontID string) (*testHasher, *stateImplTestWrapper,K_VMap) {
	testHasher := newTestHasher()
	configMap := map[string]interface{}{
		ConfigNumBuckets:             numBuckets,
		ConfigMaxGroupingAtEachLevel: maxGroupingAtEachLevel,
		ConfigHashFunction:           testHasher.getHashFunction(),
	}

	testDBWrapper.CleanDB(t)
	stateImpl := NewBucketTree(testAccontID)
	stateImpl.Initialize(configMap)
	stateImplTestWrapper := &stateImplTestWrapper{configMap, stateImpl, t}
	key_valueMap := make(K_VMap)
	return testHasher, stateImplTestWrapper, key_valueMap
}

func expectedBucketHashForTest(data ...[]string) []byte {
	return testutil.ComputeCryptoHash(expectedBucketHashContentForTest(data...))
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