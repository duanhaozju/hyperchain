// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bucket

import (
	"github.com/ToQoz/gopwt/assert"
	"github.com/hyperchain/hyperchain/common"
	"github.com/op/go-logging"
	"reflect"
	"sync"
	"testing"
)

var (
	LogOnce sync.Once
	logger  *logging.Logger
)

func NewTestLog() *logging.Logger {
	LogOnce.Do(func() {
		conf := common.NewRawConfig()
		common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
		logger = common.GetLogger(common.DEFAULT_NAMESPACE, "buckettree")
		common.SetLogLevel(common.DEFAULT_NAMESPACE, "buckettree", "NOTICE")
	})
	return logger
}

func NewTestConfig() map[string]interface{} {
	return map[string]interface{}{
		hashTableCapcity:    10009,
		aggreation:          10,
		merkleNodeCacheSize: 10000,
		bucketCacheSize:     10000,
	}
}

func NewEmptyConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestInitConfig(t *testing.T) {
	conf := initConfig(NewTestLog(), NewTestConfig())
	expect := &config{
		lowest: 5,
		levelInfo: map[int]int{
			0: 1,
			1: 2,
			2: 11,
			3: 101,
			4: 1001,
			5: 10009,
		},
	}
	assert.OK(t, reflect.DeepEqual(conf.levelInfo, expect.levelInfo), "level info should be equal")
	assert.OK(t, conf.getLowestLevel() == expect.lowest, "lowest should be equal")
	assert.OK(t, conf.getCapacity() == 10009, "capacity should be equal")

	conf = initConfig(NewTestLog(), NewEmptyConfig())
	assert.OK(t, reflect.DeepEqual(conf.levelInfo, expect.levelInfo), "level info should be equal")
	assert.OK(t, conf.getLowestLevel() == expect.lowest, "lowest should be equal")
	assert.OK(t, conf.getCapacity() == 10009, "capacity should be equal")
}

func TestComputeIndex(t *testing.T) {
	conf := initConfig(NewTestLog(), NewTestConfig())
	assert.OK(t, conf.computeIndex([]byte("helloworld"))%19 == 5, "hash index should be equal")
}

func TestComputeParentIndex(t *testing.T) {
	conf := initConfig(NewTestLog(), NewTestConfig())
	assert.OK(t, conf.computeParentIndex(0) == 0, "")
	assert.OK(t, conf.computeParentIndex(9) == 1, "")
	assert.OK(t, conf.computeParentIndex(10) == 1, "")
	assert.OK(t, conf.computeParentIndex(19) == 2, "")
}

func TestGetMaxIndex(t *testing.T) {
	conf := initConfig(NewTestLog(), NewTestConfig())
	assert.OK(t, conf.getMaxIndex(0) == 1, "")
	assert.OK(t, conf.getMaxIndex(1) == 2, "")
}
