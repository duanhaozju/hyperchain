package network_test

import (
	. "hyperchain/p2p/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
type testCase struct {
	src string
	want string
	wantport int
}

var _ = Describe("Net", func() {
	var testcases1 = []testCase{
		{"127.0.0.1:8081", "127.0.0.1",8081},
		{"localhost:8081", "localhost",8081},
		{"172.16.10.10:12121", "172.16.10.10",12121},
	}
	_ = []testCase{
		{"127.0.0.18081", "",0},
		{"hellow8081", "",0},
		{"hellow:whwe", "",0},
	}
	Describe("parse the dns item",func(){
		Context("with some correctly items",func(){
			for _,kase := range testcases1{
				It("this should be collectly",func(){
					hostname,port,err := ParseRoute(kase.src)
					Expect(hostname).To(Equal(kase.want))
					Expect(port).To(Equal(kase.wantport))
					Expect(err).To(BeNil())
			})
			}

		})
	})
})
