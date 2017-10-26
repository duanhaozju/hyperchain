package chain

import (
	"github.com/golang/protobuf/proto"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/common/service/server"
	"hyperchain/core/types"
	"sync"
)

type iChain interface {
    AddChain(namespace string, chain *memChain) error
    GetChain(namespace string) *memChain
}

// remoteMemChains remote memChains
type remoteMemChains struct {
	is   *server.InternalServer
	lock sync.RWMutex
}

func NewRemoteMemChains(is *server.InternalServer) *remoteMemChains {
	return &remoteMemChains{
		is: is,
	}
}

func (chains *remoteMemChains) AddChain(namespace string, chain *memChain) error {
	return nil
}

func (chains *remoteMemChains) GetChain(namespace string) *memChain {
	chains.lock.RLock()
	defer chains.lock.RUnlock()
	msg := &pb.IMessage{
		Type:  pb.Type_RESPONSE,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ValidationEvent,
		Payload: []byte(namespace),
	}
	s := chains.is.ServerRegistry().Namespace(namespace).Service(service.EXECUTOR)
	s.Send(msg)
	respMsg := <-s.Response()
    if respMsg.Type == pb.Type_RESPONSE || respMsg.Ok == true {
        chain := &types.MemChain{}
        proto.Unmarshal(respMsg.Payload, chain)
        memChain := &memChain{}
        memChain.data = *chain.Data
        //memChain.cpChan <- *chain.CpChan
        memChain.txDelta = chain.TxDelta
        return memChain
    } else {
        logger(namespace).Errorf("get remote chain err")
    }

	return nil
}

// ==========================================================
// Public functions that invoked by outer service
// ==========================================================

// InitializeChain inits the chains instance at first time,
// and adds a memChain with given namespace.
func InitializeRemoteChain(is *server.InternalServer, namespace string) {
	// Construct the chains only at the first time
    chainsInitOnce.Do(func() {
        chains = NewRemoteMemChains(is)
	})
}

