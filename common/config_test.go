//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewEmptyConfig(t *testing.T) {
	conf := NewRawConfig()
	conf.Set("k1", "v1")
	v := conf.GetString("k1")
	if v != "v1" {
		t.Error("config set or put error")
	}
}

func TestGetString(t *testing.T) {
	var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

	vk := "server.version"
	expect := "0.1"

	rs := conf.GetString(vk)
	if rs != expect {
		t.Errorf("GetString(%q) = %s, actual: %s", vk, rs, expect)
	}
}

func TestGetInt(t *testing.T) {
	var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

	key := "server.port"
	expect := int64(50051)

	rs := conf.GetInt64(key)
	if rs != expect {
		t.Errorf("GetInt64(%q) = %d, actual : %d", key, rs, expect)
	}
}

func TestGetDuration(t *testing.T) {
	var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

	key := "server.duration"
	expect, _ := time.ParseDuration("3s")
	rs := conf.GetDuration(key)

	if !reflect.DeepEqual(rs, expect) {
		t.Errorf("GetDuration(%q) = %v, actual: %v", key, rs, expect)
	}
}

func TestGetFloat64(t *testing.T) {
	var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

	key := "server.tls.key.value"
	expect := 12.34

	rs := conf.GetFloat64(key)
	if rs != expect {
		t.Errorf("GetFloat64(%q) = %f, actual : %f", key, rs, expect)
	}
}

func TestGetBool(t *testing.T) {
	var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

	key := "server.tls.cert.need"
	expect := false

	rs := conf.GetBool(key)
	if rs != expect {
		t.Errorf("GetBool(%q) = %t, actual: %t", key, rs, expect)
	}
}

func TestReadConfigFile(t *testing.T) {
	conf := NewConfig("/Users/wangxiaoyi/codes/go/src/hyperchain/configuration/namespaces/global/config/namespace.toml")
	conf.Print()
	fmt.Println(conf.GetStringMap("log.module"))
}

func TestConfigMerge(t *testing.T) {
	conf := NewConfig("/Users/wangxiaoyi/codes/go/src/hyperchain/configuration/namespaces/global/config/pbft.yaml")
	conf.Print()

	conf.MergeConfig("/Users/wangxiaoyi/codes/go/src/hyperchain/configuration/namespaces/global/config/namespace.toml")
	conf.Print()
}

func TestReadTomlConfigFile(t *testing.T) {
	conf := NewConfig("/Users/wangxiaoyi/codes/go/src/hyperchain/configuration/global.toml")
	conf2 := NewConfig("/Users/wangxiaoyi/codes/go/src/hyperchain/configuration/global.yaml")

	fmt.Println(conf.equals(conf2))

	conf.Print()
}
