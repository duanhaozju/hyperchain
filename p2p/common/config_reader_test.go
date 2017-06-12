package common

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/p2p/peermessage"
	"os"
	"strings"
	"testing"
)

var (
	conf        *ConfigReader
	projectPath string
)

func init() {
	env := os.Getenv("GOPATH")
	if strings.Contains(env, ":") {
		env = strings.Split(env, ":")[0]
	}
	projectPath = env + "/src/hyperchain"
	conf = NewConfigReader(projectPath + "/p2p/peerComm/test/test_peerconfig.json")

}

func TestConfigReader_GetID(t *testing.T) {
	assert.Equal(t, conf.GetID(1), 1)
	assert.Equal(t, conf.GetID(2), 2)
	assert.Equal(t, conf.GetID(3), 3)
	assert.Equal(t, conf.GetID(4), 4)
	assert.Equal(t, conf.GetID(5), 5)
	assert.Equal(t, conf.GetID(6), 6)
	assert.Equal(t, conf.GetID(7), 7)
}

func TestConfigReader_GetIP(t *testing.T) {
	assert.Equal(t, conf.GetIP(1), "127.0.0.1")
	assert.Equal(t, conf.GetIP(2), "127.0.0.1")
	assert.Equal(t, conf.GetIP(3), "127.0.0.1")
	assert.Equal(t, conf.GetIP(4), "127.0.0.1")
	assert.Equal(t, conf.GetIP(5), "127.0.0.1")
	assert.Equal(t, conf.GetIP(6), "127.0.0.1")
	assert.Equal(t, conf.GetIP(7), "127.0.0.1")
}
func TestConfigReader_GetRPCPort(t *testing.T) {
	assert.Equal(t, conf.GetRPCPort(1), 8081)
	assert.Equal(t, conf.GetRPCPort(2), 8082)
	assert.Equal(t, conf.GetRPCPort(3), 8083)
	assert.Equal(t, conf.GetRPCPort(4), 8084)
	assert.Equal(t, conf.GetRPCPort(5), 8085)
	assert.Equal(t, conf.GetRPCPort(6), 8086)
	assert.Equal(t, conf.GetRPCPort(7), 8087)
}

func TestConfigReader_GetPort(t *testing.T) {
	assert.Equal(t, conf.GetPort(1), 8001)
	assert.Equal(t, conf.GetPort(2), 8002)
	assert.Equal(t, conf.GetPort(3), 8003)
	assert.Equal(t, conf.GetPort(4), 8004)
	assert.Equal(t, conf.GetPort(5), 8005)
	assert.Equal(t, conf.GetPort(6), 8006)
	assert.Equal(t, conf.GetPort(7), 8007)
}

func TestConfigReader_GetIntroducerID(t *testing.T) {
	assert.Equal(t, conf.IntroID(), 1)
}

func TestConfigReader_GetIntroducerIP(t *testing.T) {
	assert.Equal(t, conf.IntroIP(), "127.0.0.1")
}

func TestConfigReader_GetIntroducerPort(t *testing.T) {
	assert.Equal(t, conf.IntroPort(), 8001)
}
func TestConfigReader_GetIntroducerJSONRPCPort(t *testing.T) {
	assert.Equal(t, conf.IntroJSONRPCPort(), 8081)
}

func TestConfigReader_GetLocalGRPCPort(t *testing.T) {
	assert.Equal(t, conf.LocalGRPCPort(), 8001)
}

func TestConfigReader_GetLocalID(t *testing.T) {
	assert.Equal(t, conf.LocalID(), 1)
}

func TestConfigReader_GetLocalIP(t *testing.T) {
	assert.Equal(t, conf.LocalIP(), "127.0.0.1")
}

func TestConfigReader_GetLocalJsonRPCPort(t *testing.T) {
	assert.Equal(t, conf.LocalJsonRPCPort(), 8081)
}

func TestConfigReader_IsOrigin(t *testing.T) {
	assert.Equal(t, conf.IsOrigin(), true)
}

func TestConfigReader_IsVP(t *testing.T) {
	assert.Equal(t, conf.IsVP(), false)
}

func TestConfigReader_GetMaxPeerNumber(t *testing.T) {
	assert.Equal(t, conf.MaxNum(), 7)
}

func TestConfigReader_DelNodesAndPersist(t *testing.T) {
	peerlist := make(map[string]message.PeerAddr)
	p1 := message.NewPeerAddr("127.0.0.1", 8005, 8085, 5)
	p2 := message.NewPeerAddr("127.0.0.1", 8006, 8086, 6)
	p3 := message.NewPeerAddr("127.0.0.1", 8007, 8087, 7)
	peerlist[p1.Hash] = *p1
	peerlist[p2.Hash] = *p2
	peerlist[p3.Hash] = *p3
	conf := NewConfigReader("./test/local_peerconfig.json")
	conf.DelNodesAndPersist(peerlist)
	assert.Equal(t, conf.MaxNum(), 4)
}

func TestConfigReader_AddNodesAndPersist(t *testing.T) {
	peerlist := make(map[string]message.PeerAddr)
	p1 := message.NewPeerAddr("127.0.0.1", 8005, 8085, 5)
	p2 := message.NewPeerAddr("127.0.0.1", 8006, 8086, 6)
	p3 := message.NewPeerAddr("127.0.0.1", 8007, 8087, 7)
	peerlist[p1.Hash] = *p1
	peerlist[p2.Hash] = *p2
	peerlist[p3.Hash] = *p3
	conf := NewConfigReader("./test/local_peerconfig.json")
	conf.AddNodesAndPersist(peerlist)
	assert.Equal(t, conf.MaxNum(), 7)
}
