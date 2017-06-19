package network_test

import (
	. "hyperchain/p2p/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("Server_Chat",func(){
		Context("when start server",func(){
			It("should return non-error",func(){
				s := NewServer("server_test")
				Expect(s.StartServer(50014)).To(BeNil())
				s.StopServer()
			})
		})
	})
})
