package p2p

// PeerManager provides the basic functions which supports the peer to peer
// data transfer. Those should be invoked by the higher layer.
type PeerManager interface {
	AddNode
	DeleteNode
	MsgSender
	InfoGetter
	// initialize the peerManager which is for init the local node
	Start() error
	Stop()
}

// MsgSender Send msg to others peer
type MsgSender interface {
	// broadcast information to peers
	Broadcast(payLoad []byte)

	// send a message to specific peer  UNICAST
	SendMsg(payLoad []byte, peerList []uint64)

	// random select a VP and send msg to it
	SendRandomVP(payload []byte) error

	// broadcast information to NVP peers
	BroadcastNVP(payLoad []byte) error

	// send a message to specific NVP peer (by nvp hash) UNICAST
	SendMsgNVP(payLoad []byte, nvpList []string) error
}

// AddNode
type AddNode interface {
	// update routing table when new peer's join request is accepted
	UpdateRoutingTable(payLoad []byte)
	GetLocalAddressPayload() []byte
	SetOnline()
}

// DeleteNode interface
type DeleteNode interface {
	GetLocalNodeHash() string
	GetRouterHashifDelete(hash string) (string, uint64, uint64)
	DeleteNode(hash string) error // if self {...} else{...}
	//Delete NVP node which is connect to self
	DeleteNVPNode(hash string) error
}

// InfoGetter get the peer info to manager
type InfoGetter interface {
	// Get local node id
	GetNodeId() int
	// Get current N
	GetN() int
	//get the peer information of all nodes.
	GetPeerInfo() PeerInfos
	// set
	SetPrimary(id uint64) error
	// use by new peer when join the chain dynamically only
	GetRouters() []byte

	//ISVP local node is vp node or not
	IsVP() bool
}
