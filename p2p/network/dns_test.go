package network

import (
	"testing"
	"hyperchain/p2p/utils"
	"github.com/stretchr/testify/assert"
)



func TestDNSResolver_ListHosts(t *testing.T) {
	dnsr,err := NewDNSResolver(utils.GetProjectPath() + "/p2p/test/conf/hosts.yaml");
	if err !=nil{
		t.Error("failed",err)
		t.FailNow()
	}
	// TODO should checkout the dnsr is nil or not
	dnsr.ListHosts()
}

func TestDNSResolver_ListHosts2(t *testing.T) {
	dnsr,err := NewDNSResolver(utils.GetProjectPath() + "/p2p/test/conf/hosts.yaml");
	if err !=nil{
		t.Fatal(err)
	}
	ip,err := dnsr.GetDNS("node1")
	if err !=nil{
		t.Fatal(err)
	}
	assert.Equal(t,"127.0.0.1:50012",ip)
	ip2,err := dnsr.GetDNS("node2")
	if err !=nil{
		t.Fatal(err)
	}
	assert.Equal(t,"localhost:50012",ip2)
}

