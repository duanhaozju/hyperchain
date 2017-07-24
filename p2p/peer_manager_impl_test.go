package p2p_test

import (
	. "hyperchain/p2p"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p/network"
	"github.com/spf13/viper"
	"hyperchain/p2p/utils"
	"hyperchain/manager/event"
)

var _ = Describe("PeerManagerImpl", func() {
	Describe("New Peer manager and test it", func() {
		var hypernet *network.HyperNet
		var err error
		BeforeEach(func() {
			vip := viper.New()
			vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
			vip.Set("global.p2p.retrytime", "3s")
			vip.Set("global.p2p.port",50019)
			hypernet,err = network.NewHyperNet(vip)
			//Expect(err).To(BeNil())
			//hypernet.InitServer()
			//hypernet.InitClients()
		})

		Context("new peermanager with a correctly config", func() {
			It("It should return a correct instance", func() {
				Skip("asdasdsa")
				ev := new(event.TypeMux)
				peerconf := viper.New()
				peerconf.SetConfigFile(utils.GetProjectPath()+"/p2p/test/peerconfig.yaml")
				err := peerconf.ReadInConfig()
				Expect(err).To(BeNil())
				impl ,err := NewPeerManagerImpl("test",peerconf,ev,hypernet)
				Expect(err).To(BeNil())
				impl.Listening()
			})
		})


		AfterEach(func() {
			hypernet.Stop()
		})
	})
})
