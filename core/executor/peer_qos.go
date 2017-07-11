package executor

import (
	"encoding/json"
	"github.com/op/go-logging"
	"hyperchain/common"
	"io/ioutil"
	"math"
	"os"
)

const (
	staticPeerFile = "global.configs.static_peers"
)

// Peer qos statistic implementation
type QosStat struct {
	score             map[uint64]int64
	conf              *common.Config
	latestSelected    uint64
	logger            *logging.Logger
	ctx               *ChainSyncContext
	needUpdateGenesis bool
}

// Static peer are used as pre-configured connections which are always
// has the higest priority to regard as a state update target peer.
type StaticPeer struct {
	ID            uint64 `json:"id,omitempty"`
	LocalAddress  string `json:"local_address,omitempty"`
	RemoteAddress string `json:"remote_address,omitempty"`
	Port          uint   `json:"port,omitempty"`
}

type StaticPeers struct {
	peers []StaticPeer `json:"static_peers,omitempty"`
}

func NewQos(ctx *ChainSyncContext, conf *common.Config, logger *logging.Logger) *QosStat {
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
	qosStat := &QosStat{
		conf:              conf,
		score:             score,
		logger:            logger,
		needUpdateGenesis: needUpdateGenesis,
		ctx:               ctx,
	}
	staticPeers := qosStat.ReadStaticPeer(qosStat.conf.GetString(staticPeerFile))
	logger.Debug("read static peers from configuration", staticPeers)
	for _, peer := range staticPeers {
		if _, exist := score[peer]; exist == true {
			qosStat.score[peer] += 10
		}
	}
	qosStat.PrintScoreboard()
	return qosStat
}

// SelectPeer select a best network status via selection strategy
// regard as the target peer in state update.
func (qosStat *QosStat) SelectPeer() uint64 {
	var max int64 = int64(math.MinInt64)
	var bPeer uint64
	// select other with higest score
	if qosStat.needUpdateGenesis {
		// Simple selection, use the peer has lowest genesis as target peer,
		// target peer will never change during the sync.
		if len(qosStat.ctx.PartPeers) > 0 {
			bPeer = Min(qosStat.ctx.PartPeers)
			qosStat.logger.Noticef("select peer %d from `PartPeer` collections", bPeer)
		} else {
			return 0
		}
	} else {
		for peer, score := range qosStat.score {
			if max < score {
				bPeer = peer
				max = score
			}
		}
		qosStat.logger.Noticef("select peer %d from `FullPeer` collections", bPeer)
	}
	qosStat.latestSelected = bPeer
	return bPeer
}

// FeedBack feedback state update result(success or not), to affect the peer statistic data.
func (qosStat *QosStat) FeedBack(success bool) {
	if success {
		qosStat.incScore()
	} else {
		qosStat.decScore()
	}
}

// ReadStaticPeer load preconfig peer info
func (qosStat *QosStat) ReadStaticPeer(path string) []uint64 {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if buf, err := ioutil.ReadFile(path); err == nil {
		var peers []StaticPeer
		var ret []uint64
		err := json.Unmarshal(buf, &peers)
		if err != nil {
			qosStat.logger.Error(err.Error())
		}
		for _, p := range peers {
			ret = append(ret, p.ID)
		}
		return ret
	}
	qosStat.logger.Debug("invalid static peers configuration file")
	return nil
}

func (qosStat *QosStat) incScore() {
	qosStat.logger.Debugf("increase peer %d score from %d to %d", qosStat.latestSelected, qosStat.score[qosStat.latestSelected], qosStat.score[qosStat.latestSelected]+1)
	qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] + 1
}

func (qosStat *QosStat) decScore() {
	qosStat.logger.Debugf("increase peer %d score from %d to %d", qosStat.latestSelected, qosStat.score[qosStat.latestSelected], qosStat.score[qosStat.latestSelected]-1)
	if qosStat.score[qosStat.latestSelected] > 10 {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] / 2
	} else if qosStat.score[qosStat.latestSelected] > -10 {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] - 1
	} else {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] * 2
	}
}

func (qosStat *QosStat) PrintScoreboard() {
	qosStat.logger.Debug("<====== scoreboard =======>")
	for pId, score := range qosStat.score {
		qosStat.logger.Debugf("<====== peer (id #%d), score (#%d) =======>", pId, score)
	}
	qosStat.logger.Debug("<======     end    =======>")
}

func Min(peers []PartPeer) uint64 {
	var genesis uint64 = math.MaxUint64
	var peer PartPeer
	for _, p := range peers {
		if p.Genesis < genesis {
			genesis = p.Genesis
			peer = p
		}
	}
	return peer.Id
}
