package msg

import (
	"fmt"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/payloads"
	"github.com/op/go-logging"
	"hyperchain/p2p/peerevent"
	"github.com/pkg/errors"
	"hyperchain/crypto/csprng"
)

type ClientHelloMsgHandler struct {
	shts   *hts.ServerHTS
	logger *logging.Logger
	hub    *event.TypeMux
}

func NewClientHelloHandler(shts *hts.ServerHTS, mgrhub *event.TypeMux, logger *logging.Logger) *ClientHelloMsgHandler {
	return &ClientHelloMsgHandler{
		shts:   shts,
		logger: logger,
		hub:    mgrhub,
	}
}

func (h *ClientHelloMsgHandler) Process() {
	h.logger.Info("client hello message not support stream message, so need not listen the stream message.")
}

func (h *ClientHelloMsgHandler) Teardown() {
	h.logger.Info("client hello msg not support the stream message, so needn't to be close")
}

func (h *ClientHelloMsgHandler) Receive() chan<- interface{} {
	h.logger.Info("client hello message not support stream message")
	return nil
}

func (h *ClientHelloMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	h.logger.Infof("got a client hello message %v", msg)
	rsp := &pb.Message{
		MessageType: pb.MsgType_SERVERHELLO,
		Payload:     []byte("server hello response(msg from server)"),
	}
	//got a identity payload
	id, err := payloads.IdentifyUnSerialize(msg.Payload)
	if err != nil {
		h.logger.Warningf("server hello error: %s \n",err.Error())
		return nil, err
	}
	// Key Agree
	if id.Payload == nil{
		h.logger.Error("server hello id.Payload is nil")
		return nil,errors.New("server hello id.Payload is nil")
	}
	rand,err := csprng.CSPRNG(32)
	if err != nil{

		return nil,errors.New("server hello id.Payload is nil")
	}

	cert,err := payloads.CertificateUnMarshal(id.Payload)
	r := append(cert.Rand,rand...)

	err = h.shts.KeyExchange(id.Hash,r,cert.ECert)
	if err !=nil{
		return nil,errors.New(fmt.Sprintf("cannot complete the key exchange, reason: %s",err.Error()))
	}

	if !id.IsOriginal && id.IsVP {
		//if verify passed, should notify peer manager to reverse connect to client.
		// if VP/NVP both should reverse to connect.
		fmt.Println("---==---===--->> post ev VPCONNECT")
		go h.hub.Post(peerevent.EV_VPConnect{
			Hostname: id.Hostname,
			Namespace:id.Namespace,
			ID:int(id.Id),
		})

	} else if !id.IsOriginal && !id.IsVP{
		//if is nvp h.hub.Post(peerevent.EV_NVPConnect{})
		fmt.Println("************---->> post ev NVPCONNECT")
		go h.hub.Post(peerevent.EV_NVPConnect{
			Hostname: id.Hostname,
			Namespace:id.Namespace,
			Hash:id.Hash,
		})
	}else{
		//do nothing
	}

	return rsp, nil
}
