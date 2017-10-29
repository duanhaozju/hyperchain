package p2p

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/terasum/viper"
)

var _ = Describe("P2PManager", func() {
	vip := viper.New()
	vip.SetConfigFile(utils.GetProjectPath()+"/p2p/test/global.toml")

	vip.Set(common.P2P_RETRY_TIME, "3s")
	vip.Set(common.P2P_PORT, 50019)
	vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
	vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/p2p/test/tls/tlsca.ca")
	vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.cert")
	vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.priv")

	var (
		p2pMgr    P2PManager
		p2pMgrErr error
	)

	JustBeforeEach(func() {
		err := vip.ReadInConfig()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if p2pMgr != nil {
			p2pMgr.Stop()
			ClearP2PManager()
		}
	})

	Describe("Starting up p2p manager", func() {
		BeforeEach(func() {
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/notFound.toml")
		})

		Context("with a invalid config file", func() {
			It("should return a error with a invalid p2p config", func() {

				By(fmt.Sprintf("invalid config string gloabl.hosts %s", vip.GetString(common.P2P_HOSTS)))

				p2pMgr, p2pMgrErr = GetP2PManager(vip)
				Expect(p2pMgrErr).NotTo(BeNil())
				Expect(p2pMgr).To(BeNil())
			})
		})

	})

	Describe("Starting up p2p manager", func() {
		BeforeEach(func() {
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
		})

		Context("with a valid config file", func() {

			It("should start up a global p2pManager successfully", func() {
				p2pMgr, p2pMgrErr = GetP2PManager(vip)
				Expect(p2pMgrErr).To(BeNil())
				Expect(p2pMgr).NotTo(BeNil())
			})

			Describe("getting a peer manager", func() {
				vip.Set(common.PEER_CONFIG_PATH, utils.GetProjectPath()+"/p2p/test/peerconfig.toml")
				peerConfigPath := vip.GetString(common.PEER_CONFIG_PATH)
				peerConfig := viper.New()
				peerConfig.SetConfigFile(peerConfigPath)
				delFlag := make(chan bool)

				It("should read peer config successfully", func() {
					err := peerConfig.ReadInConfig()
					Expect(err).To(BeNil())
				})

				It("should get a peer manager instance", func() {
					peerMgr, err := p2pMgr.GetPeerManager("", peerConfig, new(event.TypeMux), delFlag)
					Expect(err).To(BeNil())
					Expect(peerMgr).NotTo(BeNil())
				})

			})
		})

	})

	Describe("Create a new PeerManager", func() {
		var peerConfigPath string
		BeforeEach(func() {
			vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
			vip.Set(common.PEER_CONFIG_PATH, utils.GetProjectPath()+"/p2p/test/peerconfig.toml")
			peerConfigPath = vip.GetString(common.PEER_CONFIG_PATH)

			p2pMgr, p2pMgrErr = GetP2PManager(vip)
			Expect(p2pMgrErr).To(BeNil())
			Expect(p2pMgr).NotTo(BeNil())
		})
		It("should create PeerManager successfully", func() {
			delFlag := make(chan bool)
			peerMgr, err := GetPeerManager("", peerConfigPath, new(event.TypeMux), delFlag)
			Expect(err).To(BeNil())
			Expect(peerMgr).NotTo(BeNil())
		})

	})
})
