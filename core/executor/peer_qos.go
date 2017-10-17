// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"math"
	"sync"
)

// Peer qos statistic implementation
type QosStat struct {
	score             map[uint64]int64
	conf              *common.Config
	latestSelected    uint64
	logger            *logging.Logger
	ctx               *chainSyncContext
	needUpdateGenesis bool
	once              sync.Once
	namespace         string
}

func NewQos(ctx *chainSyncContext, conf *common.Config, namespace string, logger *logging.Logger) *QosStat {
	// assign init score
	score := make(map[uint64]int64)
	var needUpdateGenesis bool
	if len(ctx.fullPeers) == 0 {
		for _, peer := range ctx.partPeers {
			score[peer.Id] = 0
		}
		needUpdateGenesis = true
	} else {
		for _, peer := range ctx.fullPeers {
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
		namespace:         namespace,
	}
	staticPeers := qosStat.ReadStaticPeer()
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
		if len(qosStat.ctx.partPeers) > 0 {
			bPeer = Min(qosStat.ctx.partPeers)
			qosStat.logger.Noticef("select peer %d from `PartPeer` collections", bPeer)
		} else {
			return 0
		}
	} else {
		// Select peer with higest score as the best peer.
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
func (qosStat *QosStat) ReadStaticPeer() []uint64 {
	var ret []uint64
	qosStat.once.Do(func() {
		qosStat.conf.MergeConfig(common.GetPath(qosStat.namespace, qosStat.conf.GetString(common.PEER_CONFIG_PATH)))
	})
	nodes := qosStat.conf.Get("nodes").([]interface{})
	for _, item := range nodes {
		var id uint64
		node := item.(map[string]interface{})
		for key, value := range node {
			if key == "id" {
				id = uint64(value.(int64))
				ret = append(ret, id)
			}
		}
	}
	qosStat.logger.Debugf("local node static peer setting: %v", ret)
	return ret
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
