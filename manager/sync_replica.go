package manager

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/p2p/peermessage"
	"hyperchain/recovery"
	"time"
)

type ReplicaInfo struct {
	IP   string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Port int64  `protobuf:"varint,2,opt,name=Port" json:"Port,omitempty"`
	Hash string `protobuf:"bytes,3,opt,name=hash" json:"hash,omitempty"`
	ID   uint64 `protobuf:"varint,4,opt,name=ID" jsonson:"ID,omitempty"`

	LatestBlockHash []byte `protobuf:"bytes,1,opt,name=latestBlockHash,proto3" json:"latestBlockHash,omitempty"`
	ParentBlockHash []byte `proto3obuf:"bytes,2,opt,name=parentBlockHash,proto3" json:"parentBlockHash,omitempty"`
	Height          uint64 `protobuf:"varint,3,opt,name=height" json:"heightight,omitempty"`
}

func (self *ProtocolManager) SyncReplicaStatus() {
	ticker := time.NewTicker(self.syncReplicaInterval)
	for {
		select {
		case <-ticker.C:
			addr, chain := self.packReplicaStatus()
			if addr == nil || chain == nil {
				continue
			}
			status := &recovery.ReplicaStatus{
				Addr:  addr,
				Chain: chain,
			}
			payload, err := proto.Marshal(status)
			if err != nil {
				log.Error("marshal syncReplicaStatus message failed")
				continue
			}
			peers := self.Peermanager.GetAllPeers()
			var peerIds = make([]uint64, len(peers))
			for idx, peer := range peers {
				//TODO change id into int type
				peerIds[idx] = uint64(peer.PeerAddr.ID)
			}
			self.Peermanager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
			// post to self
			self.eventMux.Post(event.ReplicaStatusEvent{
				Payload: payload,
			})
		}
	}
}

func (self *ProtocolManager) RecordReplicaStatus(ev event.ReplicaStatusEvent) {
	status := &recovery.ReplicaStatus{}
	proto.Unmarshal(ev.Payload, status)
	addr := &peermessage.PeerAddress{}
	chain := &types.Chain{}
	proto.Unmarshal(status.Addr, addr)
	proto.Unmarshal(status.Chain, chain)
	replicaInfo := ReplicaInfo{
		IP: addr.IP,
		Port: int64(addr.Port),
		Hash: addr.Hash,
		ID:              uint64(addr.ID),
		LatestBlockHash: chain.LatestBlockHash,
		ParentBlockHash: chain.ParentBlockHash,
		Height:          chain.Height,
	}
	log.Debug("recv Replica Status", replicaInfo)
	self.replicaStatus.Add(addr.ID, replicaInfo)
}
func (self *ProtocolManager) packReplicaStatus() ([]byte, []byte) {
	peerAddress := self.Peermanager.GetLocalNode().GetNodeAddr()
	currentChain := core.GetChainCopy()
	// remove useless fields
	currentChain.RequiredBlockNum = 0
	currentChain.RequireBlockHash = nil
	currentChain.RecoveryNum = 0
	currentChain.CurrentTxSum = 0
	// marshal
	addrData, err := proto.Marshal(peerAddress.ToPeerAddress())
	if err != nil {
		log.Error("packReplicaStatus failed!")
		return nil, nil
	}
	chainData, err := proto.Marshal(currentChain)
	if err != nil {
		log.Error("packReplicaStatus failed!")
		return nil, nil
	}
	return addrData, chainData
}
