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
	"hyperchain/core/blockpool"
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
	blockPool   *blockpool.BlockPool
	Peermanager p2p.PeerManager

	nodeInfo  client.PeerInfos // node info ,store node status,ip,port
	consenter consensus.Consenter
	//encryption   crypto.Encryption
	AccountManager *accounts.AccountManager
	commonHash     crypto.CommonHash

	eventMux      *event.TypeMux
	txSub         event.Subscription
	newBlockSub   event.Subscription
	consensusSub  event.Subscription
	viewChangeSub event.Subscription
	respSub       event.Subscription

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

func NewProtocolManager(blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
	//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
	am *accounts.AccountManager, commonHash crypto.CommonHash) *ProtocolManager {

	manager := &ProtocolManager{

		blockPool:   blockPool,
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:   consenter,
		Peermanager: peerManager,
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
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{}, event.NewTxEvent{})
	pm.newBlockSub = pm.eventMux.Subscribe(event.CommitOrRollbackBlockEvent{}, event.ExeTxsEvent{})
	pm.syncCheckpointSub = pm.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{})
	pm.syncBlockSub = pm.eventMux.Subscribe(event.ReceiveSyncBlockEvent{})
	pm.respSub = pm.eventMux.Subscribe(event.RespInvalidTxsEvent{})
	pm.viewChangeSub = pm.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	go pm.NewBlockLoop()
	go pm.ConsensusLoop()
	go pm.syncBlockLoop()
	go pm.syncCheckpointLoop()
	go pm.respHandlerLoop()
	go pm.viewChangeLoop()
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
			if core.GetChainCopy().RecoveryNum >= blockChainInfo.Height {
				return
			}
			required := &recovery.CheckPointMessage{
				RequiredNumber: blockChainInfo.Height,
				CurrentNumber:  core.GetChainCopy().Height,
				PeerId:         UpdateStateMessage.Id,
			}

			// For Test
			// Midify the current highest block
			/*
				db, _ := hyperdb.GetLDBDatabase()
				blk, _ := core.GetBlockByNumber(db, core.GetChainCopy().Height)
				blk.BlockHash = []byte("fakehash")
				core.UpdateChain(blk, false)
			*/
			//log.Error(required.PeerId)
			core.UpdateRequire(blockChainInfo.Height, blockChainInfo.CurrentBlockHash, blockChainInfo.Height)
			// save context
			core.SetReplicas(UpdateStateMessage.Replicas)
			core.SetId(UpdateStateMessage.Id)

			payload, _ := proto.Marshal(required)
			message := &recovery.Message{
				MessageType:  recovery.Message_SYNCCHECKPOINT,
				MsgTimeStamp: time.Now().UnixNano(),
				Payload:      payload,
			}
			broadcastMsg, _ := proto.Marshal(message)
			self.Peermanager.SendMsgToPeers(broadcastMsg, UpdateStateMessage.Replicas, recovery.Message_SYNCCHECKPOINT)

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
				block, err := core.GetBlockByNumber(db, i)
				if err != nil {
					log.Warning("no required block number")
				}

				if blocks.Batch == nil {
					blocks.Batch = append(blocks.Batch, block)
				} else {
					blocks.Batch[0] = block

				}

				payload, _ := proto.Marshal(blocks)
				message := &recovery.Message{
					MessageType:  recovery.Message_SYNCBLOCK,
					MsgTimeStamp: time.Now().UnixNano(),
					Payload:      payload,
				}
				var peers []uint64
				peers = append(peers, checkpointMsg.PeerId)
				broadcastMsg, _ := proto.Marshal(message)

				self.Peermanager.SendMsgToPeers(broadcastMsg, peers, recovery.Message_SYNCBLOCK)
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
					if blocks.Batch[i].Number <= core.GetChainCopy().RequiredBlockNum {
						log.Debug("Receive Block: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
						if blocks.Batch[i].Number == core.GetChainCopy().RequiredBlockNum {
							acceptHash := blocks.Batch[i].HashBlock(self.commonHash).Bytes()
							if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(core.GetChainCopy().RequireBlockHash) {
								var tmp = blocks.Batch[i].Number - 1
								var tmpHash = blocks.Batch[i].ParentHash
								for tmp > core.GetChainCopy().Height {
									if ret, _ := core.GetBlockByNumber(db, tmp); ret != nil {
										tmp = tmp - 1
										tmpHash = ret.ParentHash
									} else {
										break
									}
								}
								core.UpdateRequire(tmp, tmpHash, core.GetChainCopy().RecoveryNum)
								log.Debug("Next Required", core.GetChainCopy().RequiredBlockNum, common.BytesToHash(core.GetChainCopy().RequireBlockHash).Hex())
								core.PutBlockTx(db, self.commonHash, blocks.Batch[i].BlockHash, blocks.Batch[i])

								// receive all block in chain
								if core.GetChainCopy().RequiredBlockNum <= core.GetChainCopy().Height {
									//如果刚好是最后一个要添加的区块
									lastBlk, _ := core.GetBlockByNumber(db, core.GetChainCopy().RequiredBlockNum+1)
									if common.Bytes2Hex(lastBlk.ParentHash) == common.Bytes2Hex(core.GetChainCopy().LatestBlockHash) {
										// execute all received block at one time
										log.Debug("Recv All blocks required, match")
										for i := core.GetChainCopy().RequiredBlockNum + 1; i <= core.GetChainCopy().RecoveryNum; i += 1 {
											blk, err := core.GetBlockByNumber(db, i)
											//originMerkleRoot := blk.MerkleRoot
											if err != nil {
												continue
											} else {
												self.blockPool.ProcessBlockInVm(blk.Transactions, nil, blk.Number)
												self.blockPool.SetDemandNumber(blk.Number + 1)
												self.blockPool.SetDemandSeqNo(blk.Number + 1)

												/*
													if bytes.Compare(blk.MerkleRoot, originMerkleRoot) != 0 {
														// stateDb has difference status
													}
												*/
											}
										}
										core.UpdateChainByBlcokNum(db, core.GetChainCopy().RecoveryNum)
										core.UpdateRequire(uint64(0), []byte{}, uint64(0))
										core.SetReplicas(nil)
										core.SetId(0)

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
										time.Sleep(2000 * time.Millisecond)
										self.consenter.RecvMsg(msgPayload)
										break
									} else {
										//如果自己链上最新区块异常,则替换,并广播节点需要的最新区块
										core.DeleteBlockByNum(db, lastBlk.Number-1)
										core.UpdateChainByBlcokNum(db, lastBlk.Number-2)
										self.broadcastDemandBlock(lastBlk.Number-1, lastBlk.ParentHash, core.GetReplicas(), core.GetId())
									}
								}
							}

						} else {
							// block with smaller height arrive early than expected
							// put into db
							if ret, _ := core.GetBlockByNumber(db, blocks.Batch[i].Number); ret != nil {
								log.Debug("Receive Duplicate Block: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
								// already exists in db
								continue
							}
							core.PutBlockTx(db, self.commonHash, blocks.Batch[i].BlockHash, blocks.Batch[i])
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

		case event.CommitOrRollbackBlockEvent:
			self.blockPool.CommitBlock(ev, self.Peermanager)

		case event.ExeTxsEvent:
			go self.blockPool.Validate(ev, self.commonHash, self.AccountManager.Encryption, self.Peermanager)

		}
	}
}

func (self *ProtocolManager) respHandlerLoop() {

	for obj := range self.respSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.RespInvalidTxsEvent:
			// receive invalid tx message, save to db
			self.blockPool.StoreInvalidResp(ev)
		}
	}
}
func (self *ProtocolManager) viewChangeLoop() {

	for obj := range self.viewChangeSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			// receive invalid tx message, save to db
			self.blockPool.ResetStatus(ev)
		case event.InformPrimaryEvent:
			//log.Notice("InformPrimaryEvent")
			self.Peermanager.SetPrimary(ev.Primary)
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
			go self.Peermanager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
			//go self.peerManager.SendMsgToPeers(ev.Payload,)
		case event.NewTxEvent:
			log.Debug("###### enter NewTxEvent")
			go self.sendMsg(ev.Payload)

		case event.ConsensusEvent:
			//call consensus module
			log.Debug("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)

		}

	}
}

func (self *ProtocolManager) sendMsg(payload []byte) {
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
	pm.Peermanager.BroadcastPeers(payload)

}

func (pm *ProtocolManager) GetNodeInfo() client.PeerInfos {
	pm.nodeInfo = pm.Peermanager.GetPeerInfo()
	log.Info("nodeInfo is ", pm.nodeInfo)
	return pm.nodeInfo

}
func (pm *ProtocolManager) broadcastDemandBlock(number uint64, hash []byte, replicas []uint64, peerId uint64) {
	required := &recovery.CheckPointMessage{
		RequiredNumber: number,
		CurrentNumber:  core.GetChainCopy().Height,
		PeerId:         peerId,
	}
	payload, _ := proto.Marshal(required)
	message := &recovery.Message{
		MessageType:  recovery.Message_SYNCSINGLE,
		MsgTimeStamp: time.Now().UnixNano(),
		Payload:      payload,
	}
	broadcastMsg, _ := proto.Marshal(message)
	pm.Peermanager.SendMsgToPeers(broadcastMsg, replicas, recovery.Message_SYNCSINGLE)
}
