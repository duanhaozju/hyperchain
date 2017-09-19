package network_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "hyperchain/p2p/network"
)

var _ = Describe("Server", func() {
	Describe("Server_Chat", func() {
		Context("when start server", func() {
			It("should return non-error", func() {
				s := NewServer("server_test", nil, nil)
				Expect(s.StartServer(":50014")).To(BeNil())
				s.StopServer()
			})
		})
	})
})
