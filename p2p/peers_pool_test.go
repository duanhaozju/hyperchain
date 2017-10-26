package p2p

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/manager/event"
	"hyperchain/p2p/utils"
	"github.com/terasum/viper"
	"hyperchain/p2p/info"
)

var _ = Describe("PeersPool", func() {
	Describe("creating a new peers pool", func() {

		var peersPool *PeersPool

		ev := new(event.TypeMux)
		peerConfigPath := utils.GetProjectPath()+"/p2p/test/peerconfig.toml"
		peerConfig := viper.New()
		peerConfig.SetConfigFile(peerConfigPath)

		Context("with a correct peer config file", func() {
			It("should create peers pool successfully", func() {
				err := peerConfig.ReadInConfig()
				Expect(err).To(BeNil())

				nodes := peerConfig.Get("nodes").([]interface{})
				pts, err := QuickParsePeerTriples("", nodes)
				Expect(err).To(BeNil())

				peersPool = NewPeersPool("", ev, pts, newPeerCnf(peerConfig))
				Expect(peersPool).NotTo(BeNil())
			})
		})

		Describe("adding a VP peer", func() {

			var (
				peer *Peer
				err error
			)

			Context("with a new vp peer", func() {
				It("should add successfully", func() {
					peer, err = NewPeer("", "node5", 5, info.NewInfo(5, "node5", ""),
						nil, nil, nil)
					Expect(err).To(BeNil())
					Expect(peer).NotTo(BeNil())

					err := peersPool.AddVPPeer(5, peer)
					Expect(err).To(BeNil())
				})
			})

			Describe("getting a not null peer pool", func() {
				Context("with one peer", func() {
					It("should return true", func() {
						notEmpty := peersPool.Ready()
						Expect(notEmpty).To(BeTrue())
					})
				})
			})

			Describe("getting all VP peers", func() {
				Context("with one peer", func() {
					It("should get a array of size 1", func() {
						vpPeers := peersPool.GetPeers()
						Expect(vpPeers).NotTo(BeNil())
						Expect(len(vpPeers)).To(BeEquivalentTo(1))
					})
				})
			})

			Describe("getting the maximum ID", func() {
				It("should return 5", func() {
					id := peersPool.MaxID()
					Expect(id).To(BeEquivalentTo(5))
				})
			})

			Describe("getting peer by hash", func() {
				It("should return a peer", func() {
					p := peersPool.GetPeerByHash(peer.info.Hash)
					Expect(p).NotTo(BeNil())
				})
			})

			Describe("getting peer by hostname", func() {
				It("should return a peer", func() {
					p, isOk := peersPool.GetPeerByHostname("node5")
					Expect(isOk).To(BeTrue())
					Expect(p).NotTo(BeNil())
				})
			})

			Describe("deleting vp peer", func() {
				var pr *Peer
				BeforeEach(func() {
					pr, _ = NewPeer("", "node6", 6, info.NewInfo(6, "node6", ""),
						nil, nil, nil)
					peersPool.AddVPPeer(6, pr)
				})

				Context("by id", func() {
					It("should delete successfully", func() {
						err := peersPool.DeleteVPPeer(6)
						Expect(err).To(BeNil())
					})
				})

				Context("by hash", func() {
					It("should delete successfully", func() {
						err := peersPool.DeleteVPPeerByHash(pr.info.Hash)
						Expect(err).To(BeNil())
					})
				})
			})

			Describe("getting the number of vp peer", func() {
				It("should return a array of size 1", func() {
					n := peersPool.GetVPNum()
					Expect(n).To(BeEquivalentTo(1))
				})
			})
		})
	})
})