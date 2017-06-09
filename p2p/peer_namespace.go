package p2p

import "github.com/pkg/errors"

var (
	ERROR_CONTAINER_NOT_FOUND = errors.New("The namespace container cannot founded.")
	ERROR_PEER_ALREADY_EXIST = errors.New("This peer already exist.")
)

type container struct {
	peers map[string]*Peer
}
func (con container) addPeer(peerID string,peer *Peer) error{
	if _,ok := con.peers[peerID];ok{
		return ERROR_PEER_ALREADY_EXIST
	}
	peer[peerID] = peer
	return nil
}

type PeerNamespace struct {
	containers map[string]*container
}

func (peerNs *PeerNamespace) getContainer(namespace string) *container {
	if container,ok := peerNs.containers[namespace];ok {
		return container
	}else{
		return nil
	}

}

//Register register the namespace and peer, bind namespace, namespace's inner id and actually peer
func (peerNs *PeerNamespace) Register(namespace string,peerID string,peer *Peer) error{
	container := peerNs.getContainer(namespace)
	if container==nil{
		return ERROR_CONTAINER_NOT_FOUND
	}
	err := container.addPeer(peerID,peer)
	return err
}