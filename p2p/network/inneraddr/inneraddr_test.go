package inneraddr

import (
	"github.com/hyperchain/hyperchain/p2p/utils"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetInnerAddr(t *testing.T) {
	t.Log(utils.GetProjectPath() + "/p2p/test/addr.toml")
	_,_, err := GetInnerAddr(utils.GetProjectPath() + "/p2p/test/addr.toml")
	if err != nil {
		t.Fatal(err)
	}
}

func TestInnerAddr_Serialize(t *testing.T) {
	addr, _, err := GetInnerAddr(utils.GetProjectPath() + "/p2p/test/addr.toml")
	if err != nil {
		t.Fatal(err)
	}
	b, e := addr.Serialize()
	if e != nil {
		t.Fatal(e)
	}
	t.Log(string(b))

	ia, e := InnerAddrUnSerialize(b)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(ia)
}

func TestInnerAddr_AddDel(t *testing.T) {
	addr, _, err := GetInnerAddr(utils.GetProjectPath() + "/p2p/test/addr.toml")
	if err != nil {
		t.Fatal(err)
	}

	addr.Add("domain5","127.0.0.1:50011")
	ipAddr := addr.Get("domain5")
	assert.Equal(t, "127.0.0.1:50011", ipAddr)

	addr.Del("domain5")
	ipAddr = addr.Get("domain5")
	assert.Equal(t, "127.0.0.1:50011", ipAddr)

	addr.Del("domain4")
	addr.Del("domain3")
	addr.Del("domain2")
	addr.Del("domain1")
	ipAddr = addr.Get("domain5")
	assert.Equal(t, "", ipAddr)
}
