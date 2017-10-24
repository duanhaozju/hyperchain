package p2p_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p/info"
	"hyperchain/p2p"
)

var _ = Describe("Peer", func() {
	Describe("creating a new peer", func() {

		var (
			peer *p2p.Peer
			err error
		)

		It("should create successfully", func() {
			peer, err = p2p.NewPeer("test", "node5", 5, info.NewInfo(5, "node5", "test"),
				nil, nil, nil)
			Expect(err).To(BeNil())
			Expect(peer).NotTo(BeNil())
		})

		Context("getting peer's id", func() {
			It("should get successfully", func() {
				id := peer.Weight()
				Expect(id).To(BeEquivalentTo(5))
			})
		})

		Context("getting peer instance", func() {
			It("should get successfully", func() {
				p := peer.Value()
				Expect(p).NotTo(BeNil())
			})

		})
	})
})