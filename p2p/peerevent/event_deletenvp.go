package peerevent

type S_DELETE_NVP struct {

	// hash of NVP peer will be deleted
	Hash string

	// hash of VP peer that NVP peer connected to.
	// NVP peer should send a message to VP peer to disconnect through p2p network
	// before NVP peer disconnect to VP peer.
	VPHash string
}
