package p2p_test

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p"
	"github.com/hyperchain/hyperchain/p2p/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/terasum/viper"
)

var _ = Describe("PeerManagerImpl", func() {

	var (
		p2pMgr     p2p.P2PManager
		p2pMgrErr  error
		peerMgr    p2p.PeerManager
		peerMgrErr error
	)

	vip := viper.New()
	vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
	vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
	vip.Set(common.P2P_RETRY_TIME, "3s")
	vip.Set(common.P2P_PORT, 50019)

	Describe("creating a new PeerManager instance", func() {

		BeforeEach(func() {
			p2pMgr, p2pMgrErr = p2p.GetP2PManager(vip)
			Expect(p2pMgrErr).To(BeNil())
			Expect(p2pMgr).NotTo(BeNil())
		})

		Context("with a correct config file", func() {
			It("should return a correct instance", func() {
				ev := new(event.TypeMux)
				delFlag := make(chan bool)

				vip.Set(common.PEER_CONFIG_PATH, utils.GetProjectPath()+"/p2p/test/peerconfig.toml")
				peerConfigPath := vip.GetString(common.PEER_CONFIG_PATH)

				peerMgr, peerMgrErr = p2p.GetPeerManager("", peerConfigPath, ev, delFlag)
				Expect(peerMgrErr).To(BeNil())
				Expect(peerMgr).NotTo(BeNil())
			})

		})
	})

	Describe("starting to testing PeerManager", func() {

		BeforeEach(func() {
			By("preparing P2PManager")
			if p2pMgr == nil {
				p2pMgr, p2pMgrErr = p2p.GetP2PManager(vip)
				Expect(p2pMgrErr).To(BeNil())
				Expect(p2pMgr).NotTo(BeNil())
				p2pMgr.Start()
			}

			By("preparing PeerManager")
			if peerMgr == nil {
				ev := new(event.TypeMux)
				delFlag := make(chan bool)

				vip.Set(common.PEER_CONFIG_PATH, utils.GetProjectPath()+"/p2p/test/peerconfig.toml")
				peerConfigPath := vip.GetString(common.PEER_CONFIG_PATH)

				peerMgr, peerMgrErr = p2p.GetPeerManager("", peerConfigPath, ev, delFlag)
				Expect(peerMgrErr).To(BeNil())
				Expect(peerMgr).NotTo(BeNil())
			}
		})

		AfterEach(func() {
			if peerMgr != nil {
				peerMgr.Stop()
			}
			if p2pMgr != nil {
				p2pMgr.Stop()
				p2p.ClearP2PManager()
			}
		})

		Context("staring up PeerManager", func() {
			It("should start up PeerManager successfully", func() {
				err := peerMgr.Start()
				Expect(err).To(BeNil())
			})
		})
	})
})
