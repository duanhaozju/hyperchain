package p2p_test

import (
	. "hyperchain/p2p"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"hyperchain/p2p/network"
	"hyperchain/p2p/utils"
	"fmt"
)

var _ = Describe("P2pManager", func() {
	Describe("Start up p2p manager", func() {
		var vip *viper.Viper

		JustBeforeEach(func() {
			vip.Set("global.p2p.retrytime", "3s")
			vip.Set("global.p2p.port",50012)

		})
		AfterEach(func() {
			err := ClearP2PManager()
			Expect(err).To(BeNil())
		})

		Context("with a valid config file", func() {
			BeforeEach(func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")

			})
			It("it should start up a global p2pManager", func() {
				_,err := GetP2PManager(vip)
				Expect(err).To(BeNil())
			})

		})

	})


	Describe("Start up p2p manager wrong", func() {
		var vip *viper.Viper

		BeforeEach(func() {
			vip = viper.New()
			vip.Set("global.p2p.retrytime", "3s")
			vip.Set("global.p2p.port",50012)
			vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/notFound.yaml")

		})

		Context("with a invalid config file", func() {

			It("it should return a error with a invalid p2p config", func() {
				By(fmt.Sprintf("invalid config string gloabl.p2p.hosts %s",vip.GetString("global.p2p.hosts")))

				_,err := GetP2PManager(vip)
				Expect(err).NotTo(BeNil())
			})
		})

	})
	Describe("Server start up", func() {
		var vip *viper.Viper

		BeforeEach(func() {
			vip = viper.New()
			vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
			vip.Set("global.p2p.retrytime", "3s")
			vip.Set("global.p2p.port",50020)
		})

		Context("With a correctly config", func() {
			It("it should start up successful, and return nil", func() {
				hypernet, err := network.NewHyperNet(vip)
				Expect(err).To(BeNil())
				err = hypernet.InitServer()
				Expect(err).To(BeNil())
			})
		})

	})

})
