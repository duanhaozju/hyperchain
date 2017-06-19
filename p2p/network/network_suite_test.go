package network_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"hyperchain/p2p/network"
	"hyperchain/p2p/message"
	"hyperchain/p2p/msg"
)

var weakTestServer *network.Server
var strongerTestServer *network.Server

func TestNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Suite")
}

var _ = BeforeSuite(func(){
	weakTestServer = network.NewServer("weakServer")
	weakTestServer.StartServer(50012)
	strongerTestServer = network.NewServer("strongerServer")
	testBlackHole := make(chan *message.Message)
	testHelloHanlder := msg.NewHelloHandler(testBlackHole)
	strongerTestServer.RegisterSlot(message.Message_HELLO,testHelloHanlder)
	strongerTestServer.StartServer(50013)
})

var _ = AfterSuite(func(){
	weakTestServer.StopServer()
	strongerTestServer.StopServer()
})
