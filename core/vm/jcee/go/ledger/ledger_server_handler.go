package jcee

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	"github.com/op/go-logging"
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
		return
	}

	exist, state := h.stateMgr.GetStateDb(key.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
		return
	}
	if valid := h.requestCheck(key.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.ReturnError()
		return
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
		return
	}

	exist, state := h.stateMgr.GetStateDb(kv.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
		return
	}
	if valid := h.requestCheck(kv.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.ReturnError()
		return
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
		return
	}
}

func (h *Handler) delete(msg *pb.Message) {

	k := &pb.Key{}
	err := proto.Unmarshal(msg.Payload, k)
	if err != nil {
		h.logger.Error(err)
		h.returnFalseResponse()
		return
	}

	exist, state := h.stateMgr.GetStateDb(k.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.returnFalseResponse()
		return
	}
	if valid := h.requestCheck(k.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.returnFalseResponse()
	}
	state.SetState(common.HexToAddress(k.Context.Cid), common.BytesToRightPaddingHash(k.K), nil, 0)
	h.returnTrueResponse()
}

func (h *Handler) batchRead(msg *pb.Message) {

	bk := &pb.BatchKey{}
	err := proto.Unmarshal(msg.Payload, bk)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
		return
	}

	exist, state := h.stateMgr.GetStateDb(bk.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
		return
	}
	if valid := h.requestCheck(bk.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.ReturnError()
		return
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
		return
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
		return
	}

	exist, state := h.stateMgr.GetStateDb(batch.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.ReturnError()
		return
	}
	if valid := h.requestCheck(batch.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.returnFalseResponse()
		return
	}
	for _, kv := range batch.Kv {
		state.SetState(common.HexToAddress(batch.Context.Cid), common.BytesToRightPaddingHash(kv.K), kv.V, 0)
	}
	paylod, err := proto.Marshal(&pb.Response{Ok: true})
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
		return
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

	r := &pb.Range{}
	err := proto.Unmarshal(msg.Payload, r)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
		return
	}

	exist, state := h.stateMgr.GetStateDb(r.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.returnFalseResponse()
		return
	}
	if valid := h.requestCheck(r.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.returnFalseResponse()
		return
	}
	start := common.BytesToRightPaddingHash(r.Start)
	limit := common.BytesToRightPaddingHash(r.End)

	iterRange := vm.IterRange{
		Start: &start,
		Limit: &limit,
	}
	iter, err := state.NewIterator(common.BytesToAddress(common.FromHex(r.Context.Cid)), &iterRange)
	if err != nil {
		h.logger.Error(err)
		h.returnFalseResponse()
		return
	}
	cnt := 0
	batchValue := &pb.BathValue{
		Id: r.Context.Txid,
	}
	for iter.Next() {
		s := make([]byte, len(iter.Value()))
		copy(s, iter.Value())
		batchValue.V = append(batchValue.V, s)
		cnt += 1
		if cnt == BatchSize {
			batchValue.HasMore = true

			paylod, err := proto.Marshal(batchValue)
			if err != nil {
				h.logger.Error(err)
				h.ReturnError()
				return
			}

			err = h.stream.Send(&pb.Message{
				Type:    pb.Message_RESPONSE,
				Payload: paylod,
			})

			if err != nil {
				h.logger.Error(err)
				return
			}

			cnt = 0
			batchValue = &pb.BathValue{
				Id: r.Context.Txid,
			}
		}
	}
	batchValue.HasMore = false
	paylod, err := proto.Marshal(batchValue)
	if err != nil {
		h.logger.Error(err)
		h.ReturnError()
		return
	}

	err = h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})

	if err != nil {
		h.logger.Error(err)
		return
	}
}

func (h *Handler) post(msg *pb.Message) {

	event := &pb.Event{}
	err := proto.Unmarshal(msg.Payload, event)
	if err != nil {
		h.logger.Error(err)
		h.returnFalseResponse()
		return
	}

	exist, state := h.stateMgr.GetStateDb(event.Context.Namespace)
	if exist == false {
		h.logger.Error(NamespaceNotExistErr)
		h.returnFalseResponse()
		return
	}
	if valid := h.requestCheck(event.Context); !valid {
		h.logger.Error(InvalidRequestErr)
		h.returnFalseResponse()
		return
	}

	var topics []common.Hash
	for _, topic := range event.Topics {
		topics = append(topics, common.BytesToHash(topic))
	}

	log := types.NewLog(common.HexToAddress(event.Context.Cid), topics, event.Body, event.Context.BlockNumber)
	state.AddLog(log)

	h.returnTrueResponse()
}

func (h *Handler) ReturnError() {
	h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: []byte(""),
	})
}

func (h *Handler) returnFalseResponse() {
	paylod, _ := proto.Marshal(&pb.Response{Ok: false})
	h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
	})
}

func (h *Handler) returnTrueResponse() {
	paylod, _ := proto.Marshal(&pb.Response{Ok: true})
	h.stream.Send(&pb.Message{
		Type:    pb.Message_RESPONSE,
		Payload: paylod,
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
