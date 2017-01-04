//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"testing"
	"os"
	"time"
	"reflect"
)

var conf = NewConfig(os.Getenv("GOPATH") + "/src/hyperchain/config/test/config_test.yaml")

func TestGetString(t *testing.T)  {
	vk := "server.version"
	expect := "0.1"

	rs := conf.GetString(vk)
	if rs != expect {
		t.Errorf("GetString(%q) = %s, actual: %s", vk, rs, expect)
	}
}

func TestGetInt(t *testing.T)  {
	key := "server.port"
	expect := int64(50051)

	rs := conf.GetInt64(key)
	if rs != expect {
		t.Errorf("GetInt64(%q) = %d, actual : %d", key, rs, expect)
	}
}

func TestGetDuration(t *testing.T)  {
	key := "server.duration"
	expect, _ := time.ParseDuration("3s")
	rs := conf.GetDuration(key)

	if !reflect.DeepEqual(rs, expect) {
		t.Errorf("GetDuration(%q) = %v, actual: %v", key, rs, expect)
	}
}

func TestGetFloat64(t *testing.T) {
	key := "server.tls.key.value"
	expect := 12.34

	rs := conf.GetFloat64(key)
	if rs != expect{
		t.Errorf("GetFloat64(%q) = %f, actual : %f", key, rs, expect)
	}
}

func TestGetBool(t *testing.T) {
	key := "server.tls.cert.need"
	expect := false

	rs := conf.GetBool(key)
	if rs != expect{
		t.Errorf("GetBool(%q) = %t, actual: %t", key, rs, expect)
	}
}