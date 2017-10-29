package p2p

import (
	"github.com/hyperchain/hyperchain/p2p/info"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/terasum/viper"
	"github.com/hyperchain/hyperchain/p2p/utils"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/p2p/network"
	"github.com/hyperchain/hyperchain/p2p/hts"
	"github.com/hyperchain/hyperchain/p2p/hts/secimpl"
	"fmt"
	"github.com/hyperchain/hyperchain/manager/event"
	"time"
)

type helloHandler struct {}

func NewHelloHandler() *helloHandler{
	return &helloHandler{}
}

func (hanlder *helloHandler) Receive() chan<- interface{} { return make(chan interface{})}
func (hanlder *helloHandler) Process() {}
func (hanlder *helloHandler) Teardown() {}
func (hanlder *helloHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	return  pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("execute it")), nil
}

var _ = Describe("Peer", func() {
	Describe("creating a new peer", func() {

		var (
			peer *Peer
			err  error
		)

		vip := viper.New()
		vip.SetConfigFile(utils.GetProjectPath()+"/p2p/test/global.toml")
		vip.ReadInConfig()
		vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
		vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
		vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/p2p/test/tls/tlsca.ca")
		vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.cert")
		vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.priv")

		// initialize hypernet
		hypernet, err := network.NewHyperNet(vip, "hyperchain")
		if err != nil {
			panic(fmt.Sprintf("NewHyperNet error: %v", err))
		}
		if err = hypernet.InitServer(); err != nil {
			panic(fmt.Sprintf("InitServer error: %v", err))
		}
		if err = hypernet.InitClients(); err != nil {
			panic(fmt.Sprintf("InitClients error: %v", err))
		}

		// register handler
		hypernet.RegisterHandler("test", pb.MsgType_CLIENTHELLO, NewHelloHandler())
		hypernet.RegisterHandler("test", pb.MsgType_SERVERHELLO, NewHelloHandler())
		hypernet.RegisterHandler("test", pb.MsgType_CLIENTREJECT, NewHelloHandler())

		// initialize hts
		nsCnfigPath := utils.GetProjectPath()+"/p2p/test/namespace.toml"
		hts, err := hts.NewHTS("", secimpl.NewSecuritySelector(nsCnfigPath), nsCnfigPath)
		if err != nil {
			panic(fmt.Sprintf("NewHTS error: %v", err))
		}

		// initialize hts client
		chts, err := hts.GetAClientHTS()
		if err != nil {
			panic(fmt.Sprintf("GetAClientHTS error: %v", err))
		}

		Context("without others peer information", func() {
			It("should create successfully", func() {
				Expect(err).To(BeNil())

				peer, err = NewPeer("", "node1", 1, info.NewInfo(1, "node1", "test"),
					hypernet, chts, new(event.TypeMux))
				Expect(err).To(BeNil())
				Expect(peer).NotTo(BeNil())
			})
		})

		Context("with others peer information", func() {
			It("should create successfully", func() {
				err := hypernet.Connect("node2")
				Expect(err).To(BeNil())

				peer, err = NewPeer("", "node1", 1, info.NewInfo(3, "node3", "test"),
					hypernet, chts, new(event.TypeMux))
					time.Sleep(time.Second * 1)
				Expect(err).To(BeNil())
				Expect(peer).NotTo(BeNil())
			})
		})

		Context("to get peer id", func() {
			It("should get successfully", func() {
				id := peer.Weight()
				Expect(id).To(BeEquivalentTo(1))
			})
		})

		Context("to get peer instance", func() {
			It("should get successfully", func() {
				p := peer.Value()
				Expect(p).NotTo(BeNil())
			})
		})

		Context("to send greeting message", func() {
			It("should send successfully", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's greeting"))
				resp, err := peer.Greeting(msg)
				Expect(err).To(BeNil())
				Expect(resp.Payload).Should(Equal([]byte("execute it")))
			})
		})

		//Context("to send whisper message", func() {
		//	It("should send successfully", func() {
		//		msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's whisper"))
		//		resp, err := peer.Whisper(msg)
		//		Expect(err).To(BeNil())
		//		Expect(resp.Payload).Should(Equal([]byte("execute it")))
		//	})
		//})

		Context("to serialize peer information", func() {
			var p []byte
			It("should serialize successfully", func() {
				p = peer.Serialize()
				Expect(p).NotTo(BeNil())
			})

			Context("to deserialize peer information", func() {
				It("should deserialize failed", func() {
					hostname, namespace, hash, err := PeerDeSerialize(p)
					Expect(err).To(BeNil())
					Expect(namespace).To(Equal(""))
					Expect(hostname).To(Equal("node1"))
					Expect(hash).To(Equal(peer.info.GetHash()))
				})

			})
		})

	})
})
