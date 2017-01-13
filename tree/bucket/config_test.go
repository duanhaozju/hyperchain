package bucket

import (
	"hyperchain/tree/bucket/testutil"
	"testing"
)

func TestConfigInit(t *testing.T) {
	/*configs := viper.GetStringMap("ledger.state.dataStructure.configs")
	t.Logf("Configs loaded from yaml = %#v", configs)
	testDBWrapper.CleanDB(t)
	stateImpl := NewBucketTree("testAccountAddr")
	stateImpl.Initialize(configs)
	testutil.AssertEquals(t, conf.getNumBucketsAtLowestLevel(), configs[ConfigNumBuckets])
	testutil.AssertEquals(t, conf.getMaxGroupingAtEachLevel(), configs[ConfigMaxGroupingAtEachLevel])*/
}

func TestConfig(t *testing.T) {
	testConf := newConfig(26, 2, fnvHash)
	t.Logf("conf.levelToNumBucketsMap: [%#v]", testConf.levelToNumBucketsMap)
	testutil.AssertEquals(t, testConf.getLowestLevel(), 5)
	testutil.AssertEquals(t, testConf.getNumBuckets(0), 1)
	testutil.AssertEquals(t, testConf.getNumBuckets(1), 2)
	testutil.AssertEquals(t, testConf.getNumBuckets(2), 4)
	testutil.AssertEquals(t, testConf.getNumBuckets(3), 7)
	testutil.AssertEquals(t, testConf.getNumBuckets(4), 13)
	testutil.AssertEquals(t, testConf.getNumBuckets(5), 26)

	testutil.AssertEquals(t, testConf.computeParentBucketNumber(25), 13)
	testutil.AssertEquals(t, testConf.computeParentBucketNumber(9), 5)
	testutil.AssertEquals(t, testConf.computeParentBucketNumber(10), 5)

	testConf = newConfig(26, 3, fnvHash)
	t.Logf("conf.levelToNumBucketsMap: [%#v]", testConf.levelToNumBucketsMap)
	testutil.AssertEquals(t, testConf.getLowestLevel(), 3)
	testutil.AssertEquals(t, testConf.getNumBuckets(0), 1)
	testutil.AssertEquals(t, testConf.getNumBuckets(1), 3)
	testutil.AssertEquals(t, testConf.getNumBuckets(2), 9)
	testutil.AssertEquals(t, testConf.getNumBuckets(3), 26)

	testutil.AssertEquals(t, testConf.computeParentBucketNumber(24), 8)
	testutil.AssertEquals(t, testConf.computeParentBucketNumber(25), 9)
}
