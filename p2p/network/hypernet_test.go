package network_test

import (
	. "hyperchain/p2p/network"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p/utils"
	"github.com/spf13/viper"
	"time"
)

var _ = Describe("Hypernet", func() {
	Describe("new a hypernet instance", func() {
		var vip *viper.Viper
		Context("with a valid viper instance",func(){
			BeforeEach(func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
				vip.Set("global.p2p.retrytime", "3s")
				vip.Set("global.p2p.port",50012)
			})
			It("it should return nil err", func() {
				_, err := NewHyperNet(vip)
				Expect(err).To(BeNil())
			})

			It("retry will covery goroutine code", func() {
				_, err := NewHyperNet(vip)
				Expect(err).To(BeNil())
				<- time.After(5 * time.Second)
			})

			It("retry will cover the nil time duration branch", func() {
				vip.Set("global.p2p.retrytime", "")
				_, err := NewHyperNet(vip)
				Expect(err).NotTo(BeNil())
			})
		})

		Context("With a invalid viper instance", func() {
			It("It should return a error",func() {
				_,err := NewHyperNet(nil)
				Expect(err).NotTo(BeNil())
			})
			It("It should return a error, because invalid config path",func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/wrong.yaml")
				_,err := NewHyperNet(vip)
				Expect(err).NotTo(BeNil())
			})
		})

	})

	Describe("new a hypernet instance and init server", func() {
		var (
			hypernet *HyperNet
			err error
			vip *viper.Viper
		)
		BeforeEach(func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
				vip.Set("global.p2p.retrytime", "3s")
				vip.Set("global.p2p.port",50018)
			})
		JustBeforeEach(func() {
			hypernet,err = NewHyperNet(vip)
			Expect(err).To(BeNil())
		})

		Context("With valid config", func() {
			It("It should start a valid server, and listening at port 50018", func() {
				err = hypernet.InitServer()
				Expect(err).To(BeNil())
			})
		})

		Context("with a invalid config, wrong port", func() {
			BeforeEach(func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
				vip.Set("global.p2p.retrytime", "3s")
				vip.Set("global.p2p.port",0)
			})
			It("It should return a err when init server", func() {
				err = hypernet.InitServer()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("new a valid hypernet, test connect", func() {
		var hypernet *HyperNet
		var vip *viper.Viper
		var err error
		JustBeforeEach(func() {
			hypernet,err  = NewHyperNet(vip)
			Expect(err).To(BeNil())
			err = hypernet.InitServer()
			Expect(err).To(BeNil())
		})
		Context("", func() {
			BeforeEach(func() {
				vip = viper.New()
				vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
				vip.Set("global.p2p.retrytime", "3s")
				vip.Set("global.p2p.port",50019)
			})
			It("it should connect to all hostss", func() {
				err := hypernet.InitClients()
				Expect(err).To(BeNil())
			})
		})
	})
})
