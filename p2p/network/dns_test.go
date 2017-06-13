package network

import (
	"testing"
	"hyperchain/p2p/utils"
	"github.com/magiconair/properties/assert"
)



func TestDNSResolver_ListHosts(t *testing.T) {
	dnsr,err := NewDNSResolver(utils.GetProjectPath() + "/src/hyperchain/configuration/namespaces/hosts.yaml");
	if err !=nil{
		t.Error("failed")
		t.Error(err)
	}
	// TODO should checkout the dnsr is nil or not
	dnsr.ListHosts()
}

func TestDNSResolver_ListHosts2(t *testing.T) {
	dnsr,err := NewDNSResolver(utils.GetProjectPath() + "/src/hyperchain/configuration/namespaces/hosts.yaml");
	if err !=nil{
		t.Error(err)
		t.Fail()
	}
	// TODO should checkout the dnsr is nil or not
	hostname,port,err := ParseRoute("node1:8081")
	if err != nil{
		t.Error(err)
		t.Fail()
	}
	ip,err := dnsr.GetDNS(hostname)
	if err !=nil{
		t.Error(err)
		t.Fail()
	}
	assert.Equal(t,ip,"127.0.0.1")
	assert.Equal(t,port,8081)

}

