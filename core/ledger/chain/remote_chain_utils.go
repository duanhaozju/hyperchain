package chain

import (
	"github.com/golang/protobuf/proto"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/common/service/server"
	"hyperchain/core/types"
	"sync"
    "strconv"
)

type iChain interface {
    AddChain(namespace string, chain *memChain) error
    GetChain(namespace string, checkpoint bool) *memChain
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

func (chains *remoteMemChains) GetChain(namespace string, checkpoint bool) *memChain {
	chains.lock.RLock()
	defer chains.lock.RUnlock()
	msg := &pb.IMessage{
		Type:  pb.Type_RESPONSE,
		From:  pb.FROM_EVENTHUB,
		Payload: []byte(namespace + "," + strconv.FormatBool(checkpoint)),
	}
	s := chains.is.ServerRegistry().Namespace(namespace).Service(service.EXECUTOR)
	s.Send(msg)
	respMsg := <-s.Response()
    if respMsg.Type == pb.Type_RESPONSE || respMsg.Ok == true {
        chain := &types.MemChain{}
        err := proto.Unmarshal(respMsg.Payload, chain)
        if err != nil {
            logger(namespace).Criticalf("MemChain unmarshal err: %v", err)
            return nil
        }
        memChain := &memChain{
            data: types.Chain{},
            cpChan: make(chan types.Chain),
        }
        memChain.data = *chain.Data
        go func() {
            if chain.RemoteChan != nil {
                memChain.cpChan <- *chain.RemoteChan
            }
        }()
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

