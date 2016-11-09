// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-29
package manager

import (
	"hyperchain/consensus"
	"hyperchain/core/blockpool"

	"hyperchain/crypto"
	"hyperchain/p2p"

	"hyperchain/accounts"
	"hyperchain/protos"

	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/event"
	"time"
)

// init protocol manager params and start
func New(eventMux *event.TypeMux, blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, consenter consensus.Consenter,
//encryption crypto.Encryption, commonHash crypto.CommonHash,path string, nodeId int) (error) {
am *accounts.AccountManager, commonHash crypto.CommonHash, nodeId int, syncReplicaInterval time.Duration, syncReplica bool, expired chan bool, expiredTime time.Time) *ProtocolManager {

	aliveChan := make(chan bool)
	go peerManager.Start(aliveChan, eventMux)

	//wait for all peer are connected
	select {
	case <-aliveChan:
		{

			//protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, fetcher, consenter, encryption, commonHash)
			protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, expired, expiredTime)
			protocolManager.Start()
			// consensusEvent NegotiateView
			negoView := &protos.Message{
				Type:      protos.Message_NEGOTIATE_VIEW,
				Timestamp: time.Now().UnixNano(),
				Payload:   nil,
				Id:        0,
			}
			msg, err := proto.Marshal(negoView)
			if err != nil {
				fmt.Println("nego view start")
			}
			eventMux.Post(event.ConsensusEvent{
				Payload: msg,
			})

			//start server
			return protocolManager

		}
	}

	//return nil
}
