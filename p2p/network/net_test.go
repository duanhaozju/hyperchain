package network_test

import (
	. "github.com/hyperchain/hyperchain/p2p/network"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testCase struct {
	src      string
	want     string
	wantport int
}

func TestParseRoute(t *testing.T) {
	Convey("Net", t, func() {
		var testcases1 = []testCase{
			{"127.0.0.1:8081", "127.0.0.1", 8081},
			{"localhost:8081", "localhost", 8081},
			{"172.16.10.10:12121", "172.16.10.10", 12121},
		}
		_ = []testCase{
			{"127.0.0.18081", "", 0},
			{"hellow8081", "", 0},
			{"hellow:whwe", "", 0},
		}
		Convey("parse the dns item", func() {
			Convey("with some correctly items", func() {
				for _, kase := range testcases1 {
					hostname, port, err := ParseRoute(kase.src)
					So(hostname, ShouldEqual, kase.want)
					So(port, ShouldEqual, kase.wantport)
					So(err, ShouldBeNil)
				}

			})
		})
	})
}
