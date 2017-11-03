//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func getTestConfig() *Config {
	testConfigPath := "/src/github.com/hyperchain/hyperchain/common/testhelper/resources/global.toml"
	return NewConfig(os.Getenv("GOPATH") + testConfigPath)
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
	expect := 100
	conf := getTestConfig()
	assert.NotEqual(t, expect, conf.GetInt("consensus.rbft.batchsize"), "This config file should not contain this value")

	conf.MergeConfig(os.Getenv("GOPATH") + "/src/hyperchain/common/testhelper/resources/namespace.toml")
	assert.Equal(t, expect, conf.GetInt("consensus.rbft.batchsize"), "This config file should contain this value after merging")
}

func TestReadTomlConfigFile(t *testing.T) {
	conf := getTestConfig()
	conf2 := NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/common/testhelper/resources/namespace.toml")

	assert.NotEqual(t, true, conf.equals(conf2), "These two config file are not the same")
}
