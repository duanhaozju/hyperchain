package executor

import (
	"hyperchain/common"
	"math"
	"io/ioutil"
	"os"
	"encoding/json"
	"github.com/op/go-logging"
)
const (
	staticPeerFile = "global.configs.static_peers"
)
type Oracle struct {
	score             map[uint64]int64
	conf              *common.Config
	latestSelected    uint64
	logger            *logging.Logger
	ctx               *ChainSyncContext
	needUpdateGenesis bool
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

func NewOracle(ctx *ChainSyncContext, conf *common.Config, logger *logging.Logger) *Oracle {
	// assign init score
	score := make(map[uint64]int64)
	var needUpdateGenesis bool
	if len(ctx.FullPeers) == 0 {
		for _, peer := range ctx.PartPeers {
			score[peer.Id] = 0
		}
		needUpdateGenesis = true
	} else {
		for _, peer := range ctx.FullPeers {
			score[peer] = 0
		}
		needUpdateGenesis = false
	}
	oracle := &Oracle{
		conf:              conf,
		score:             score,
		logger:            logger,
		needUpdateGenesis: needUpdateGenesis,
	}
	staticPeers := oracle.ReadStaticPeer(oracle.conf.GetString(staticPeerFile))
	logger.Debug("read static peers from configuration", staticPeers)
	for _, peer := range staticPeers {
		if _, exist := score[peer]; exist == true {
			oracle.score[peer] += 10
		}
	}
	oracle.PrintScoreboard()
	return oracle
}
func (oracle *Oracle) SelectPeer() uint64 {
	var max int64 = int64(math.MinInt64)
	var bPeer uint64
	// select other with higest score
	if oracle.needUpdateGenesis {
		// TODO fake selection
		// TODO ask @Rongjialei fix this
		bPeer = oracle.ctx.PartPeers[0].Id
		oracle.logger.Noticef("select peer %d from `PartPeer` collections", bPeer)
	} else {
		for peer, score := range oracle.score {
			if max < score {
				bPeer = peer
				max = score
			}
		}
		oracle.logger.Noticef("select peer %d from `FullPeer` collections", bPeer)
	}
	oracle.latestSelected = bPeer
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

func (oracle *Oracle) PrintScoreboard() {
	oracle.logger.Notice("<====== scoreboard =======>")
	for pId, score := range oracle.score {
		oracle.logger.Noticef("<====== peer (id #%d), score (#%d) =======>", pId, score)
	}
	oracle.logger.Notice("<======     end    =======>")
}
