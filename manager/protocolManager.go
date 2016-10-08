// implement ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-31
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/p2p/peer"
	"hyperchain/protos"
	"hyperchain/recovery"
	"sync"
	"time"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("manager")
}

type ProtocolManager struct {
	serverPort  int
	blockPool   *core.BlockPool
	fetcher     *core.Fetcher
	peerManager p2p.PeerManager

	nodeInfo  client.PeerInfos // node info ,store node status,ip,port
	consenter consensus.Consenter
	//encryption   crypto.Encryption
	AccountManager *accounts.AccountManager
	commonHash     crypto.CommonHash

	noMorePeers  chan struct{}
	eventMux     *event.TypeMux
	txSub        event.Subscription
	newBlockSub  event.Subscription
	consensusSub event.Subscription

	aLiveSub event.Subscription

	syncCheckpointSub event.Subscription

	syncBlockSub event.Subscription
	quitSync     chan struct{}

	wg sync.WaitGroup
}
type NodeManager struct {
	peerManager p2p.PeerManager
}

var eventMuxAll *event.TypeMux

func NewProtocolManager(blockPool *core.BlockPool, peerManager p2p.PeerManager, eventMux *event.TypeMux, fetcher *core.Fetcher, consenter consensus.Consenter,
	//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
	am *accounts.AccountManager, commonHash crypto.CommonHash) *ProtocolManager {
	log.Debug("enter parotocol manager")

	manager := &ProtocolManager{

		blockPool:   blockPool,
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:   consenter,
		peerManager: peerManager,
		fetcher:     fetcher,
		//encryption:encryption,
		AccountManager: am,
		commonHash:     commonHash,
	}
	manager.nodeInfo = make(client.PeerInfos, 0, 1000)
	eventMuxAll = eventMux
	return manager
}

func GetEventObject() *event.TypeMux {
	return eventMuxAll
}

// start listen new block msg and consensus msg
func (pm *ProtocolManager) Start() {

	pm.wg.Add(1)
	go pm.fetcher.Start()
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{},event.BroadcastConsensusEvent{}, event.NewTxEvent{})
	pm.newBlockSub = pm.eventMux.Subscribe(event.NewBlockEvent{})
	pm.syncCheckpointSub = pm.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{})
	pm.syncBlockSub = pm.eventMux.Subscribe(event.ReceiveSyncBlockEvent{})
	go pm.NewBlockLoop()
	go pm.ConsensusLoop()
	go pm.syncBlockLoop()
	go pm.syncCheckpointLoop()

	pm.wg.Wait()

}
func (self *ProtocolManager) syncCheckpointLoop() {
	self.wg.Add(-1)
	for obj := range self.syncCheckpointSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.SendCheckpointSyncEvent:
			/*

				receive request  from the consensus module required block and send to  outer peers
			*/
			UpdateStateMessage := &protos.UpdateStateMessage{}
			proto.Unmarshal(ev.Payload, UpdateStateMessage)

			blockChainInfo := &protos.BlockchainInfo{}
			proto.Unmarshal(UpdateStateMessage.TargetId, blockChainInfo)

			required := &recovery.CheckPointMessage{
				RequiredNumber: blockChainInfo.Height,
				CurrentNumber:  core.GetChainCopy().Height,
				PeerId:         UpdateStateMessage.Id,
			}
			//log.Error(required.PeerId)
			core.UpdateRequire(blockChainInfo.Height, blockChainInfo.CurrentBlockHash, blockChainInfo.Height)

			payload, _ := proto.Marshal(required)
			message := &recovery.Message{
				MessageType:  recovery.Message_SYNCCHECKPOINT,
				MsgTimeStamp: time.Now().UnixNano(),
				Payload:      payload,
			}
			broadcastMsg, _ := proto.Marshal(message)
			self.peerManager.SendMsgToPeers(broadcastMsg, UpdateStateMessage.Replicas, recovery.Message_SYNCCHECKPOINT)

		case event.StateUpdateEvent:
			/*
				get required block from db and send to outer peers
			*/

			receiveMessage := &recovery.Message{}
			proto.Unmarshal(ev.Payload, receiveMessage)

			checkpointMsg := &recovery.CheckPointMessage{}
			proto.Unmarshal(receiveMessage.Payload, checkpointMsg)

			db, _ := hyperdb.GetLDBDatabase()

			blocks := &types.Blocks{}
			for i := checkpointMsg.RequiredNumber; i > checkpointMsg.CurrentNumber; i -= 1 {
				//for i := checkpointMsg.CurrentNumber + 1; i <= checkpointMsg.RequiredNumber; i += 1 {
				block, err := core.GetBlockByNumber(db, i)
				if err != nil {
					log.Warning("no required block number")
				}

				if blocks.Batch == nil {
					blocks.Batch = append(blocks.Batch, block)
				} else {
					blocks.Batch[0] = block

				}

				//blocks.Batch=
				//blocks.Batch = append(blocks.Batch, block)

				payload, _ := proto.Marshal(blocks)
				message := &recovery.Message{
					MessageType:  recovery.Message_SYNCBLOCK,
					MsgTimeStamp: time.Now().UnixNano(),
					Payload:      payload,
				}
				var peers []uint64
				peers = append(peers, checkpointMsg.PeerId)
				broadcastMsg, _ := proto.Marshal(message)

				self.peerManager.SendMsgToPeers(broadcastMsg, peers, recovery.Message_SYNCBLOCK)
			}

		}
	}
}

func (self *ProtocolManager) syncBlockLoop() {

	for obj := range self.syncBlockSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.ReceiveSyncBlockEvent:
			/*
				receive block from outer peers
			*/

			if core.GetChainCopy().RequiredBlockNum != 0 {

				message := &recovery.Message{}
				proto.Unmarshal(ev.Payload, message)
				blocks := &types.Blocks{}
				proto.Unmarshal(message.Payload, blocks)
				db, _ := hyperdb.GetLDBDatabase()

				for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {

					if blocks.Batch[i].Number == core.GetChainCopy().RequiredBlockNum {

						acceptHash := blocks.Batch[i].HashBlock(self.commonHash).Bytes()
						//todo compare receive blockHash and acceptHash
						if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(core.GetChainCopy().RequireBlockHash) {

							core.UpdateRequire(blocks.Batch[i].Number-1, blocks.Batch[i].ParentHash, core.GetChainCopy().RecoveryNum)

							//core.PutBlock(db, blocks.Batch[i].BlockHash, blocks.Batch[i])

							//core.PutBlock(db, blocks.Batch[i].BlockHash, blocks.Batch[i])
							core.PutBlockTx(db, self.commonHash,blocks.Batch[i].BlockHash, blocks.Batch[i])

							// receive all block in chain
							if blocks.Batch[i].Number<=core.GetChainCopy().Height+1{
								//如果刚好是最后一个要添加的区块
							if common.Bytes2Hex(blocks.Batch[i].ParentHash) == common.Bytes2Hex(core.GetChainCopy().LatestBlockHash) {
								core.UpdateChainByBlcokNum(db, core.GetChainCopy().RecoveryNum)

								core.UpdateRequire(uint64(0), []byte{}, uint64(0))
								payload := &protos.StateUpdatedMessage{
									SeqNo: core.GetChainCopy().Height,
								}
								msg, _ := proto.Marshal(payload)
								msgSend := &protos.Message{
									Type:      protos.Message_STATE_UPDATED,
									Payload:   msg,
									Timestamp: time.Now().UnixNano(),
									Id:        1,
								}

								msgPayload, err := proto.Marshal(msgSend)
								if err != nil {
									log.Error(err)
								}

								self.consenter.RecvMsg(msgPayload)
								break
							} else{
								//如果自己链上最新区块异常,则替换,并广播节点需要的最新区块
								//deleteBlock(db,blocks.Batch[i].Number-1)
								core.UpdateChainByBlcokNum(db, blocks.Batch[i].Number-2)
								//broadcastDemandBlock(blocks.Batch[i].Number-1,replica,msg)
								//core.UpdateChainByViewChange(blocks.Batch[i].Number-1,blocks.Batch[i].ParentHash)


							}
							}
						}
					}

				}

			}
		}
	}
}

// listen block msg
func (self *ProtocolManager) NewBlockLoop() {

	for obj := range self.newBlockSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.NewBlockEvent:
			//accept msg from consensus module
			//commit block into block pool

			log.Info("write block success")

			self.commitNewBlock(ev.Payload, ev.CommitTime)
			//self.fetcher.Enqueue(ev.Payload)

		}
	}
}

//listen consensus msg
func (self *ProtocolManager) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.BroadcastConsensusEvent:
			log.Debug("######enter broadcast")

			go self.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:

			var peers []uint64
			peers = append(peers, ev.PeerId)
			self.peerManager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
			//go self.peerManager.SendMsgToPeers(ev.Payload,)
		case event.NewTxEvent:

			go self.sendMsg(ev.Payload)

		case event.ConsensusEvent:
			//call consensus module
			log.Debug("###### enter ConsensusEvent")
			//logger.GetLogger().Println("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)
		case event.ExeTxsEvent:
			self.blockPool.ExecTxs(ev.SequenceNum,ev.Transactions)
		case event.CommitOrRollbackBlockEvent:
			self.blockPool.CommitOrRollbackBlockEvent(ev.SequenceNum,
				ev.Transactions,ev.Timestamp,ev.CommitTime,ev.CommitStatus)
		}


	}
}

func (self *ProtocolManager) sendMsg(payload []byte) {
	//Todo sign tx
	//payLoad := self.transformTx(payload)
	//if payLoad == nil {
	//	//log.Fatal("payLoad nil")
	//	log.Error("payLoad nil")
	//	return
	//}
	msg := &protos.Message{
		Type:    protos.Message_TRANSACTION,
		Payload: payload,
		//Payload: payLoad,
		Timestamp: time.Now().UnixNano(),
		Id:        0,
	}
	msgSend, _ := proto.Marshal(msg)
	self.consenter.RecvMsg(msgSend)

}

// Broadcast consensus msg to a batch of peers not knowing about it
func (pm *ProtocolManager) BroadcastConsensus(payload []byte) {
	log.Debug("begin call broadcast")
	pm.peerManager.BroadcastPeers(payload)

}

//receive tx from web,sign it and marshal it,then give it to consensus module
func (pm *ProtocolManager) transformTx(payload []byte) []byte {

	//var transaction types.Transaction
	transaction := &types.Transaction{}
	//decode tx
	proto.Unmarshal(payload, transaction)
	//hash tx
	h := transaction.SighHash(pm.commonHash)
	addrHex := string(transaction.From)
	addr := common.HexToAddress(addrHex)

	sign, err := pm.AccountManager.SignWithPassphrase(addr, h[:], "123")
	//sign, err := pm.accountManager.Sign(addr, h[:])
	if err != nil {
		log.Error(err)

		return nil
	}
	transaction.Signature = sign
	//encode tx
	payLoad, err := proto.Marshal(transaction)
	if err != nil {
		return nil
	}
	return payLoad

}

// add new block into block pool
func (pm *ProtocolManager) commitNewBlock(payload []byte, commitTime int64) {

	msgList := &protos.ExeMessage{}
	proto.Unmarshal(payload, msgList)
	block := new(types.Block)
	for _, item := range msgList.Batch {
		tx := &types.Transaction{}

		proto.Unmarshal(item.Payload, tx)

		block.Transactions = append(block.Transactions, tx)
	}
	block.Timestamp = msgList.Timestamp
	//block.CommitTime =commitTime

	block.Number = msgList.No

	log.Info("now is ", msgList.No)
	pm.blockPool.AddBlock(block, pm.commonHash, commitTime)
	//core.WriteBlock(*block)

}

func (pm *ProtocolManager) GetNodeInfo() client.PeerInfos {
	pm.nodeInfo = pm.peerManager.GetPeerInfos()
	log.Info("nodeInfo is ", pm.nodeInfo)
	return pm.nodeInfo

}
