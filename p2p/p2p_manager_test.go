package p2p_test

import (
	"github.com/spf13/viper"
	"hyperchain/p2p/network"
	"hyperchain/p2p/utils"
	"fmt"
	"hyperchain/common"
)

var _ = Describe("P2pManager", func() {
	Describe("Start up p2p manager", func() {
		var vip *viper.Viper

		JustBeforeEach(func() {
			vip.Set(common.P2P_RETRY_TIME, "3s")
			vip.Set(common.P2P_PORT, 50019)

		})
		AfterEach(func() {
			err := ClearP2PManager()
			Expect(err).To(BeNil())
		})

		Context("with a valid config file", func() {
			BeforeEach(func() {
				vip = viper.New()
				vip.Set(common.P2P_HOSTS, utils.GetProjectPath() + "/p2p/test/hosts.yaml")

			})
			It("it should start up a global p2pManager", func() {

				Skip("hahah")
				_, err := GetP2PManager(vip)
				Expect(err).To(BeNil())
			})

		})

	})

	Describe("Start up p2p manager wrong", func() {
		var vip *viper.Viper

		BeforeEach(func() {
			vip = viper.New()
			vip.Set(common.P2P_RETRY_TIME, "3s")
			vip.Set(common.P2P_PORT, 50019)
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath() + "/p2p/test/notFound.yaml")

		})

		Context("with a invalid config file", func() {
			It("it should return a error with a invalid p2p config", func() {

				Skip("hahah")
				By(fmt.Sprintf("invalid config string gloabl.p2p.hosts %s", vip.GetString(common.P2P_HOSTS)))
				_, err := GetP2PManager(vip)
				Expect(err).NotTo(BeNil())
			})
		})

	})
	Describe("Server start up", func() {
		var vip *viper.Viper

		BeforeEach(func() {
			vip = viper.New()
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath() + "/p2p/test/hosts.yaml")
			vip.Set(common.P2P_RETRY_TIME, "3s")
			vip.Set(common.P2P_PORT, 50019)
		})

		Context("With a correctly config", func() {
			It("it should start up successful, and return nil", func() {

				Skip("hahah")
				hypernet, err := network.NewHyperNet(vip)
				Expect(err).To(BeNil())
				err = hypernet.InitServer()
				Expect(err).To(BeNil())
				hypernet.Stop()
			})
		})

	})

})
