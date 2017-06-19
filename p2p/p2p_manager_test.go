package p2p_test

import (
	. "hyperchain/p2p"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"hyperchain/p2p/network"
	"hyperchain/p2p/utils"
)

var _ = Describe("P2pManager", func() {
	Describe("When start  up p2p manager", func() {
		It("it should start up a global p2pManager", func() {
			//TODO with a vaild config file path
			err := StartUpP2PManager(utils.GetProjectPath()+"/p2p/test/global.yaml")
			Expect(err).To(BeNil())
		})
		It("it should return a panic with a invalid p2p config", func() {
			err := StartUpP2PManager("")
			Expect(err).NotTo(BeNil())
		})

	})

	BeforeEach(func() {
		// should start up p2p manager first
		// this will run before every It
	})

	Describe("When Server start up", func() {
		It("it should start up successful, and return nil", func() {
			vip := viper.New()
			vip.SetConfigFile(utils.GetProjectPath() + "/p2p/test/global.yaml")
			err := vip.ReadInConfig()
			Expect(err).To(BeNil())
			vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
			hypernet, err := network.NewHyperNet(vip)
			Expect(err).To(BeNil())
			err = hypernet.InitServer(50012)
			Expect(err).To(BeNil())
		})
	})

	Describe("When client list was been parsed",func(){
		It("Every client connected to nodes correctly...", func() {

		})

	})
})
