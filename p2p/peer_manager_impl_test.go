package p2p

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/terasum/viper"
	"time"
)

var _ = Describe("PeerManagerImpl", func() {

	var (
		p2pMgr     P2PManager
		p2pMgrErr  error
		peerMgr    PeerManager
		peerMgrErr error
	)

	vip := viper.New()
	vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
	vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
	vip.Set(common.P2P_RETRY_TIME, "3s")
	vip.Set(common.P2P_PORT, 50019)

	Describe("creating a new PeerManager instance", func() {
		Context("with a correct config file", func() {
			It("should create successfully", func() {
				p2pMgr, p2pMgrErr = GetP2PManager(vip)
				Expect(p2pMgrErr).To(BeNil())
				Expect(p2pMgr).NotTo(BeNil())

				ev := new(event.TypeMux)
				delFlag := make(chan bool)

				vip.Set(common.PEER_CONFIG_PATH, utils.GetProjectPath()+"/p2p/test/peerconfig.toml")
				peerConfigPath := vip.GetString(common.PEER_CONFIG_PATH)

				peerMgr, peerMgrErr = GetPeerManager("", peerConfigPath, ev, delFlag)
				Expect(peerMgrErr).To(BeNil())
				Expect(peerMgr).NotTo(BeNil())
			})
		})

		Context("to start PeerManager", func() {
			It("should start successfully", func() {
				err := peerMgr.Start()
				Expect(err).To(BeNil())
				time.Sleep(time.Second * 5)
			})
		})

		Context("to send message", func() {
			It("should send successfully", func() {
				peerMgr.SendMsg([]byte("hello"), []uint64{1})
				time.Sleep(time.Second * 3)
			})
		})

		Context("to broadcast message", func() {
			It("should broadcat successfully", func() {
				peerMgr.Broadcast([]byte("hello"))
			})
		})

		Context("to stop PeerManager", func() {
			It("should stop successfully", func() {
				if peerMgr != nil {
					peerMgr.Stop()
				}
				if p2pMgr != nil {
					p2pMgr.Stop()
					ClearP2PManager()
				}
			})
		})
	})
})
