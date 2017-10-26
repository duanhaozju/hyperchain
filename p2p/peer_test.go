package p2p_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p"
	"hyperchain/p2p/info"
	//pb "hyperchain/p2p/message"
)

var _ = Describe("Peer", func() {
	Describe("creating a new peer", func() {

		var (
			peer *p2p.Peer
			err  error
		)

		Context("with self information", func() {
			It("should create successfully", func() {
				peer, err = p2p.NewPeer("test", "node5", 5, info.NewInfo(5, "node5", "test"),
					nil, nil, nil)
				Expect(err).To(BeNil())
				Expect(peer).NotTo(BeNil())
			})
		})

		//Context("with others peer information", func() {
		//	It("should create successfully", func() {
		//		nsCnfigPath := utils.GetProjectPath() + "/p2p/test/namespace.toml"
		//		hts, err := hts.NewHTS("", secimpl.NewSecuritySelector(nsCnfigPath), nsCnfigPath)
		//		Expect(err).To(BeNil())
		//
		//		chts, err := hts.GetAClientHTS()
		//		Expect(err).To(BeNil())
		//
		//		globalConfigPath := utils.GetProjectPath()+"/configuration/global.toml"
		//		vip := viper.New()
		//		vip.SetConfigFile(globalConfigPath)
		//		vip.ReadInConfig()
		//		vip.Set(common.P2P_PORT, 50019)
		//		vip.Set(common.P2P_RETRY_TIME, "3s")
		//		vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/configuration/hosts.toml")
		//		vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/configuration/addr.toml")
		//		vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/configuration/tls/tlsca.ca")
		//		vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/configuration/tls/tls_peer1.cert")
		//		vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/configuration/tls/tls_peer1.priv")
		//		net, err := network.NewHyperNet(vip, "hyperchaincn")
		//		Expect(err).To(BeNil())
		//
		//		err = net.Connect("node4")
		//		Expect(err).To(BeNil())
		//
		//		peer, err = p2p.NewPeer("test", "node4", 4, info.NewInfo(3, "node3", "test"),
		//			net, chts, new(event.TypeMux))
		//		Expect(err).To(BeNil())
		//		Expect(peer).NotTo(BeNil())
		//	})
		//})

		Context("and then getting peer id", func() {
			It("should get successfully", func() {
				id := peer.Weight()
				Expect(id).To(BeEquivalentTo(5))
			})
		})

		Context("and then getting peer instance", func() {
			It("should get successfully", func() {
				p := peer.Value()
				Expect(p).NotTo(BeNil())
			})
		})

		//Context("sending a greeting message", func() {
		//	msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("hello"))
		//	peer.Greeting(msg)
		//})
	})
})
