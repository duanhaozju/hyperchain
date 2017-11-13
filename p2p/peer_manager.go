package p2p

// PeerManager provides the basic functions which supports the peer to peer
// data transfer and node operation. Those should be invoked by the higher layer.
type PeerManager interface {
	AddNode
	DeleteNode
	MsgSender
	InfoGetter
	Start() error
	Stop()
}

// MsgSender sends message to other peers.
type MsgSender interface {

	// Broadcast broadcasts message to VP peers.
	Broadcast(payLoad []byte)

	// SendMsg sends a message to specific VP peer.(UNICAST)
	SendMsg(payLoad []byte, peerList []uint64)

	// SendRandomVP sends a message to a random VP.
	SendRandomVP(payload []byte) error

	// BroadcastNVP broadcasts message to NVP peers.
	BroadcastNVP(payLoad []byte) error

	// SendMsgNVP sends a message to specific NVP peer (by nvp hash).(UNICAST)
	SendMsgNVP(payLoad []byte, nvpList []string) error
}

// AddNode specifies some methods should be invoked when a new node will join.
type AddNode interface {

	// UpdateRoutingTable updates routing table when a new request to join is accepted.
	UpdateRoutingTable(payLoad []byte)

	// GetLocalAddressPayload returns the serialization information for the new node.
	GetLocalAddressPayload() []byte

	// SetOnline represents the node starts up successfully.
	SetOnline()
}

// DeleteNode specifies some methods should be invoked when a existing node will be deleted.
type DeleteNode interface {

	// GetLocalNodeHash returns local node hash.
	GetLocalNodeHash() string

	// GetRouterHashifDelete returns routing table hash after deleting specific peer.
	GetRouterHashifDelete(hash string) (string, uint64, uint64)

	// DeleteNode delete the specific hash node.
	DeleteNode(hash string) error

	// DeleteNVPNode deletes NVP node which connects to current VP node.
	DeleteNVPNode(hash string) error
}

// InfoGetter specifies some methods to get node/peer information.
type InfoGetter interface {

	// GetNodeId returns local node id.
	GetNodeId() int

	// GetN returns the number of connected node.
	GetN() int

	// GetPeerInfo returns all the peer information of local node, including itself.
	GetPeerInfo() PeerInfos

	// SetPrimary sets the specific node as the primary node.
	SetPrimary(id uint64) error

	// GetRouters is used by new peer when join the chain dynamically only.
	GetRouters() []byte

	// IsVP returns whether local node is VP node or not.
	IsVP() bool
}
