package network

import (
	"github.com/hyperchain/hyperchain/p2p/utils"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDNSResolver_ListHosts(t *testing.T) {
	Convey("Dns Reslover list all dns items", t, func() {
		dnsr, err := NewDNSResolver(utils.GetProjectPath() + "/p2p/test/conf/hosts.yaml")
		So(err, ShouldBeNil)
		dnsList := dnsr.ListHosts()
		So(len(dnsList), ShouldEqual, 4)
	})

}

func TestDNSResolver_ListHosts2(t *testing.T) {
	Convey("Parse DNS Item", t, func() {
		dnsr, err1 := NewDNSResolver(utils.GetProjectPath() + "/p2p/test/conf/hosts.yaml")
		So(err1, ShouldBeNil)
		ip, err2 := dnsr.GetDNS("node1")
		So(err2, ShouldBeNil)
		So(ip, ShouldEqual, "127.0.0.1:50012")
		ip2, err3 := dnsr.GetDNS("node2")
		So(err3, ShouldBeNil)
		So(ip2, ShouldEqual, "localhost:50012")
	})

}
