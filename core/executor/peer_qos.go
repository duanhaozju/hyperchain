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

// Peer qos (Quality of Service) statistic implementation
type QosStat struct {
	score             map[uint64]int64  // Track the quality score of each peer
	conf              *common.Config    // Refers a configuration reader
	latestSelected    uint64            // Track the last selected peer
	logger            *logging.Logger   // Refer the logger
	ctx               *chainSyncContext // Refer the chain synchronization context
	needUpdateGenesis bool              // Track if genesis need to be updated
	once              sync.Once         // Sync variable that control the once using
	namespace         string            // Refer the namespace
}

// newQos creates the qos manager of peers, inits the original score of each peer.
func newQos(ctx *chainSyncContext, conf *common.Config, namespace string, logger *logging.Logger) *QosStat {
	// Assign init scores
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

	// Read the static config the peer from file, and plus 10 to the init score
	staticPeers := qosStat.readStaticPeer()
	for _, peer := range staticPeers {
		if _, exist := score[peer]; exist == true {
			qosStat.score[peer] += 10
		}
	}
	qosStat.printScoreboard()

	return qosStat
}

// selectPeer select a best network status via selection strategy
// regard as the target peer in state update.
func (qosStat *QosStat) selectPeer() uint64 {
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

// feedBack gives a feedback result of state update(success or not), to affect the peer statistic data.
func (qosStat *QosStat) feedBack(success bool) {
	if success {
		qosStat.incScore()
	} else {
		qosStat.decScore()
	}
}

// readStaticPeer loads the peer info set in config file.
func (qosStat *QosStat) readStaticPeer() []uint64 {
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

// incScore increases the score of the latest selected peer.
func (qosStat *QosStat) incScore() {
	qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] + 1
	qosStat.logger.Debugf("increase peer %d score from %d to %d", qosStat.latestSelected, qosStat.score[qosStat.latestSelected], qosStat.score[qosStat.latestSelected]+1)
}

// decScore decreases the score of the latest selected peer according to different rules.
func (qosStat *QosStat) decScore() {
	origin := qosStat.score[qosStat.latestSelected]
	if qosStat.score[qosStat.latestSelected] > 10 {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] / 2
	} else if qosStat.score[qosStat.latestSelected] > -10 {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] - 1
	} else {
		qosStat.score[qosStat.latestSelected] = qosStat.score[qosStat.latestSelected] * 2
	}
	qosStat.logger.Debugf("decrease peer %d score from %d to %d", qosStat.latestSelected, origin, qosStat.score[qosStat.latestSelected])
}

// printScoreboard print the scores of each peer.
func (qosStat *QosStat) printScoreboard() {
	qosStat.logger.Debug("<====== scoreboard =======>")
	for pId, score := range qosStat.score {
		qosStat.logger.Debugf("<====== peer (id #%d), score (#%d) =======>", pId, score)
	}
	qosStat.logger.Debug("<======     end    =======>")
}

// Min finds the peer whose genesis is the smallest,
// which means that peer has more unsnapped blocks for state update.
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
