package p2p_test

var _ = Describe("PeersPool", func() {
	Describe("when create a new peers pool and add a new peer", func() {
		It("this should return nil err and create success!", func() {
			peersPool := NewPeersPool("test")
			err := peersPool.AddVPPeer(1, nil)
			Expect(err).To(BeNil())
		})
	})
})
