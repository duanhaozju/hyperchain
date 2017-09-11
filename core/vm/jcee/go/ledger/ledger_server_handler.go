package jcee

import (
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	pb "hyperchain/core/vm/jcee/protos"
	"strings"
)

type Handler struct {
	requests chan *pb.Message
	stream   pb.Contract_RegisterServer
	stateMgr *StateManager
	close    chan struct{}

	logger *logging.Logger
}

func NewHandler(requsts chan *pb.Message, stream pb.Contract_RegisterServer) *Handler {
	return &Handler{
		requests: requsts,
		stream:   stream,
		close:    make(chan struct{}),
	}
}

//handle, receive message and handle it.
func (h *Handler) handle() {
	for {
		select {
		case <-h.close:
			return
		case msg := <-h.requests:
			h.handleMsg(msg)
		}
	}
}

func (h *Handler) handleMsg(msg *pb.Message) {
	switch msg.Type {
	case pb.Message_GET:
		h.get(msg)
	case pb.Message_PUT:
		h.put(msg)
	case pb.Message_DELETE:
		h.delete(msg)
	case pb.Message_BATCH_READ:
		h.batchRead(msg)
	case pb.Message_BATCH_WRITE:
		h.batchWrite(msg)
	case pb.Message_RANGE_QUERY:
		h.rangeQuery(msg)
	case pb.Message_POST_EVENT:
		h.post(msg)
	case pb.Message_UNDEFINED:
		//undefined
	}
}

func (h *Handler) get(msg *pb.Message) {
	key := &pb.Key{}
	err := proto.Unmarshal(msg.Payload, key)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	exist, state := h.stateMgr.GetStateDb(key.Context.Namespace)
	if exist == false {
		//return nil, NamespaceNotExistErr
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
	}
	if valid := h.requestCheck(key.Context); !valid {
		//return nil, InvalidRequestErr
		h.logger.Error(InvalidRequestErr)
		h.ReturnError()
	}
	_, value := state.GetState(common.HexToAddress(key.Context.Cid), common.BytesToRightPaddingHash(key.K))
	v := &pb.Value{
		Id: key.Context.Txid,
		V:  value,
	}

	paylod, err := proto.Marshal(v)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}
	err = h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})

	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) put(msg *pb.Message) {
	kv := &pb.KeyValue{}
	err := proto.Unmarshal(msg.Payload, kv)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	exist, state := h.stateMgr.GetStateDb(kv.Context.Namespace)
	if exist == false {
		//return &pb.Response{Ok: false}, NamespaceNotExistErr
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
	}
	if valid := h.requestCheck(kv.Context); !valid {
		//return &pb.Response{Ok: false}, InvalidRequestErr
		h.logger.Error(InvalidRequestErr)
		h.ReturnError()
	}
	state.SetState(common.HexToAddress(kv.Context.Cid), common.BytesToRightPaddingHash(kv.K), kv.V, 0)

	paylod, err := proto.Marshal(&pb.Response{Ok: true})
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}
	err = h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})

	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) delete(msg *pb.Message) {
	//key *pb.Key
}

func (h *Handler) batchRead(msg *pb.Message) {

	bk := &pb.BatchKey{}
	err := proto.Unmarshal(msg.Payload, bk)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	exist, state := h.stateMgr.GetStateDb(bk.Context.Namespace)
	if exist == false {
		//return nil, NamespaceNotExistErr
		h.ReturnError()
	}
	if valid := h.requestCheck(bk.Context); !valid {
		//return nil, InvalidRequestErr
		h.ReturnError()
	}
	response := &pb.BathValue{}
	for _, key := range bk.K {
		exist, value := state.GetState(common.HexToAddress(bk.Context.Cid), common.BytesToRightPaddingHash(key))
		if exist == true {
			response.V = append(response.V, value)
		} else {
			response.V = append(response.V, nil)
		}
	}
	response.HasMore = false
	response.Id = bk.Context.Txid

	paylod, err := proto.Marshal(response)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	err = h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})

	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) batchWrite(msg *pb.Message) {

	batch := &pb.BatchKV{}
	err := proto.Unmarshal(msg.Payload, batch)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	exist, state := h.stateMgr.GetStateDb(batch.Context.Namespace)
	if exist == false {
		//return &pb.Response{Ok: false}, NamespaceNotExistErr
		h.ReturnError()
	}
	if valid := h.requestCheck(batch.Context); !valid {
		//return &pb.Response{Ok: false}, InvalidRequestErr
	}
	for _, kv := range batch.Kv {
		state.SetState(common.HexToAddress(batch.Context.Cid), common.BytesToRightPaddingHash(kv.K), kv.V, 0)
	}
	paylod, err := proto.Marshal(&pb.Response{Ok: true})
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
	}

	err = h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})

	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) rangeQuery(msg *pb.Message) {

}

func (h *Handler) post(msg *pb.Message) {

}

func (h *Handler) ReturnError() {
	h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: []byte(""),
	})
}

func (h *Handler) requestCheck(ctx *pb.LedgerContext) bool {
	exist, state := h.stateMgr.GetStateDb(ctx.Namespace)
	if exist == false {
		return false
	}
	if strings.Compare(ctx.Txid, state.GetCurrentTxHash().Hex()) == 0 {
		return true
	} else {
		return false
	}
}
