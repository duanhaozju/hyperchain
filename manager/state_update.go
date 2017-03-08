package manager

import (
	"hyperchain/event"
)

// SendSyncRequest - send synchronization request to other nodes.
func (self *ProtocolManager) SendSyncRequest(ev event.SendCheckpointSyncEvent) {
	self.executor.SendSyncRequest(ev)
}

// ReceiveSyncRequest - receive synchronization request from some node, send back request blocks.
func (self *ProtocolManager) ReceiveSyncRequest(ev event.StateUpdateEvent) {
	self.executor.ReceiveSyncRequest(ev)
}

// ReceiveSyncBlocks - receive request blocks from others.
func (self *ProtocolManager) ReceiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	self.executor.ReceiveSyncBlocks(ev)
}