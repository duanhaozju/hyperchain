package executor

func (executor *Executor) SyncReplicaInfo() {
	if executor.GetSyncReplicaEnable() {
		go executor.sendReplicaInfo()
	} else {
		log.Debug("the switch of sync replica doesn't been turn on")
	}
}

func (executor *Executor) sendReplicaInfo() {
	//interval := executor.GetSyncReplicaInterval()
	//ticker := time.NewTicker(interval)
	//for {
	//	select {
	//	case <-ticker.C:
	//		addr, chain := self.packReplicaStatus()
	//		if addr == nil || chain == nil {
	//			continue
	//		}
	//		status := &recovery.ReplicaStatus{
	//			Addr:  addr,
	//			Chain: chain,
	//		}
	//		payload, err := proto.Marshal(status)
	//		if err != nil {
	//			log.Error("marshal syncReplicaStatus message failed")
	//			continue
	//		}
	//		peers := self.Peermanager.GetAllPeers()
	//		var peerIds = make([]uint64, len(peers))
	//		for idx, peer := range peers {
	//			//TODO change id into int type
	//			peerIds[idx] = uint64(peer.PeerAddr.ID)
	//		}
	//		self.Peermanager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
	//	// post to self
	//		self.eventMux.Post(event.ReplicaStatusEvent{
	//			Payload: payload,
	//		})
	//	}
	//}
}

func (executor *Executor) receiveReplicaInfo() {

}

func (executor *Executor) packingReplicaInfo() {
	//chain := edb.GetChainCopy(executor.namespace)

}