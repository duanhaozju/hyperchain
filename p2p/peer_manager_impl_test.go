package p2p_test

import (
	"github.com/spf13/viper"
	"hyperchain/common"
	"hyperchain/manager/event"
	"hyperchain/p2p/network"
	"hyperchain/p2p/utils"
)

var _ = Describe("PeerManagerImpl", func() {
	Describe("New Peer manager and test it", func() {
		var hypernet *network.HyperNet
		var err error
		BeforeEach(func() {
			vip := viper.New()
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.yaml")
			vip.Set(common.P2P_RETRY_TIME, "3s")
			vip.Set(common.P2P_PORT, 50019)
			hypernet, err = network.NewHyperNet(vip)
			//Expect(err).To(BeNil())
			//hypernet.InitServer()
			//hypernet.InitClients()
		})

		Context("new peermanager with a correctly config", func() {
			It("It should return a correct instance", func() {
				Skip("asdasdsa")
				ev := new(event.TypeMux)
				peerconf := viper.New()
				peerconf.SetConfigFile(utils.GetProjectPath() + "/p2p/test/peerconfig.yaml")
				err := peerconf.ReadInConfig()
				Expect(err).To(BeNil())
				impl, err := NewPeerManagerImpl("test", peerconf, ev, hypernet)
				Expect(err).To(BeNil())
				impl.Listening()
			})
		})

		AfterEach(func() {
			hypernet.Stop()
		})
	})
})
