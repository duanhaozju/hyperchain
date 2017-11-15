//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"
)

var testResourcePrefix = path.Join(os.Getenv("GOPATH"), "/src/github.com/hyperchain/hyperchain/common/testhelper/resources")

func getTestConfig() *Config {
	return NewConfig(path.Join(testResourcePrefix, "global.toml"))
}

func TestNewEmptyConfig(t *testing.T) {
	conf := NewRawConfig()
	conf.Set("k1", "v1")
	v := conf.GetString("k1")
	assert.Equal(t, "v1", v, "config set or put error")
}

func TestGetString(t *testing.T) {
	var conf = getTestConfig()

	key1 := "title"
	expect1 := "Hyperchain global configurations"
	rs1 := conf.GetString(key1)
	assert.Equal(t, expect1, rs1, fmt.Sprintf("GetString(%q) = %s, actual: %s", key1, rs1, expect1))

	key2 := "log.module.p2p"
	expect2 := "DEBUG"
	rs2 := conf.GetString(key2)
	assert.Equal(t, expect2, rs2, fmt.Sprintf("GetString(%q) = %s, actual: %s", key2, rs2, expect2))
}

func TestGetInt(t *testing.T) {
	var conf = getTestConfig()

	key := "port.jsonrpc"
	expect := int64(8081)

	rs := conf.GetInt64(key)
	assert.Equal(t, expect, rs, fmt.Sprintf("GetInt64(%q) = %d, actual : %d", key, rs, expect))
}

func TestGetDuration(t *testing.T) {
	var conf = getTestConfig()

	key := "p2p.keepAliveDuration"
	expect, _ := time.ParseDuration("3s")
	rs := conf.GetDuration(key)

	assert.Equal(t, expect, rs, fmt.Sprintf("GetDuration(%q) = %v, actual: %v", key, rs, expect))
}

func TestGetBool(t *testing.T) {
	var conf = getTestConfig()

	key := "admin.check"
	expect := false

	rs := conf.GetBool(key)
	assert.Equal(t, expect, rs, fmt.Sprintf("GetBool(%q) = %t, actual: %t", key, rs, expect))
}

func TestReadConfigFile(t *testing.T) {
	conf := getTestConfig()
	conf.Print()
	fmt.Println(conf.GetStringMap("log.module"))
}

func TestConfigMerge(t *testing.T) {
	conf := getTestConfig()
	conf.MergeConfig(path.Join(testResourcePrefix, "namespace.toml"))
	assert.Equal(t, 100, conf.GetInt("consensus.rbft.batchsize"), "This config file should contain this value after merging")
}
