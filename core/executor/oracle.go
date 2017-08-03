package executor

import (
	"hyperchain/common"
	"math"
	"io/ioutil"
	"os"
	"encoding/json"
	"github.com/op/go-logging"
	"sync"
)
type Oracle struct {
	peers           []uint64
	staticPeers     []uint64
	score           map[uint64]int64
	conf            *common.Config
	latestSelected  uint64
	logger          *logging.Logger
	once            sync.Once
}

type StaticPeer struct {
	ID               uint64  `json:"id,omitempty"`
	LocalAddress     string  `json:"local_address,omitempty"`
	RemoteAddress    string  `json:"remote_address,omitempty"`
	Port             uint    `json:"port,omitempty"`
}

type StaticPeers struct {
	peers     []StaticPeer `json:"static_peers,omitempty"`
}

func NewOracle(peers []uint64, conf *common.Config, logger *logging.Logger) *Oracle {
	// assign init score
	score := make(map[uint64]int64)
	for _, peer := range peers {
		score[peer] = 0
	}
	oracle := &Oracle{
		peers:       peers,
		conf:        conf,
		score:       score,
		logger:      logger,
	}
	staticPeers := oracle.ReadPeerIds()
	logger.Debug("read static peers from configuration", staticPeers)
	for _, peer := range staticPeers {
		if contains(peers, peer) {
			oracle.staticPeers = append(oracle.staticPeers, peer)
		}
	}
	logger.Debug("static peers", oracle.staticPeers)
	return oracle
}
func (oracle *Oracle) SelectPeer() uint64 {
	var max int64 = int64(math.MinInt64)
	var bPeer uint64
	// select static peer first
	// if a static peer's score less than -5, drop it
	for _, p := range oracle.staticPeers {
		if oracle.score[p] > -5 {
			oracle.latestSelected = p
			oracle.logger.Debugf("select static peer %d to send sync req", p)
			return p
		}
	}
	// select other with higest score
	for peer, score := range oracle.score {
		if max < score {
			bPeer = peer
			max = score
		}
	}
	oracle.latestSelected = bPeer
	oracle.logger.Debugf("select peer %d to send sync req", bPeer)
	return bPeer
}
func (oracle *Oracle) FeedBack(success bool) {
	if success {
		oracle.incScore()
	} else {
		oracle.decScore()
	}
}
func (oracle *Oracle) ReadStaticPeer(path string) []uint64 {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if buf, err := ioutil.ReadFile(path); err == nil {
		var peers []StaticPeer
		var ret   []uint64
		err := json.Unmarshal(buf, &peers)
		if err != nil {
			oracle.logger.Error(err.Error())
		}
		for _, p := range peers {
			ret = append(ret, p.ID)
		}
		return ret
	}
	oracle.logger.Debug("invalid static peers configuration file")
	return nil
}

func (oracle *Oracle) ReadPeerIds() []uint64 {
	var ret []uint64
	oracle.once.Do(func() {
		oracle.conf.MergeConfig(oracle.conf.GetString(common.PEER_CONFIG_PATH))
	})

	nodes := oracle.conf.Get("nodes").([]interface{})
	for _, item := range nodes{
		var id uint64
		node := item.(map[string]interface{})
		for key, value := range node{
			if key == "id" {
				id  = uint64(value.(int64))
				ret = append(ret, id)
			}
		}
	}
	return ret
}

func contains(s []uint64, e uint64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func (oracle *Oracle) incScore() {
	oracle.logger.Debugf("increase peer %d score from %d to %d", oracle.latestSelected, oracle.score[oracle.latestSelected], oracle.score[oracle.latestSelected] + 1)
	oracle.score[oracle.latestSelected] = oracle.score[oracle.latestSelected] + 1
}
func (oracle *Oracle) decScore() {
	oracle.logger.Debugf("increase peer %d score from %d to %d", oracle.latestSelected, oracle.score[oracle.latestSelected], oracle.score[oracle.latestSelected] - 1)
	if oracle.score[oracle.latestSelected] > 10 {
		oracle.score[oracle.latestSelected] = oracle.score[oracle.latestSelected] / 2
	} else if oracle.score[oracle.latestSelected] > -10 {
		oracle.score[oracle.latestSelected] = oracle.score[oracle.latestSelected] - 1
	} else {
		oracle.score[oracle.latestSelected] = oracle.score[oracle.latestSelected] * 2
	}
}
