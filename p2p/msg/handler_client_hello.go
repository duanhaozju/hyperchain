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
	"hyperchain/p2p/info"
)

type ClientHelloMsgHandler struct {
	shts   *hts.ServerHTS
	logger *logging.Logger
	hub    *event.TypeMux
	local *info.Info
	isOrigin bool
}

func NewClientHelloHandler(shts *hts.ServerHTS, mgrhub *event.TypeMux,localInfo *info.Info,isorigin bool, logger *logging.Logger) *ClientHelloMsgHandler {
	return &ClientHelloMsgHandler{
		shts:   shts,
		logger: logger,
		hub:    mgrhub,
		local: localInfo,
		isOrigin:isorigin,
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
	h.logger.Debugf("got a client hello message %v", msg)
	// SERVER HELLO response
	data := []byte("hyperchain")
	esign, err := h.shts.CG.ESign(data)
	if err !=nil{
		return nil,err
	}
	rsign, err := h.shts.CG.RSign(data)
	if err !=nil{
		return nil,err
	}
	serverRand,err := csprng.CSPRNG(32)
	if err !=nil{
		return nil,err
	}
	certpayload,err := payloads.NewCertificate(data,h.shts.CG.GetECert(),esign,h.shts.CG.GetRCert(),rsign,serverRand)
	if err !=nil{
		return nil,err
	}
	// peer should
	identify := payloads.NewIdentify(h.local.IsVP,h.isOrigin,h.local.GetNameSpace(), h.local.Hostname, h.local.Id,certpayload)
	payload, err := identify.Serialize()
	if err != nil {
		return nil,err
	}

	rsp := pb.NewMsg(pb.MsgType_SERVERHELLO, payload)

	// NEGOTIATE shared key
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
	clientRand,err := csprng.CSPRNG(32)
	if err != nil{

		return nil,errors.New("server hello id.Payload is nil")
	}

	cert,err := payloads.CertificateUnMarshal(id.Payload)
	r := append(cert.Rand,clientRand...)

	err = h.shts.KeyExchange(id.Hash,r,cert.ECert)
	if err !=nil{
		h.logger.Errorf("handler_client_hello.go 70 %s \n",err.Error())
		return nil,errors.New(fmt.Sprintf("cannot complete the key exchange, reason: %s",err.Error()))
	}

	if !id.IsOriginal && id.IsVP {
		//if verify passed, should notify peer manager to reverse connect to client.
		// if VP/NVP both should reverse to connect.
		h.logger.Info("post ev VPCONNECT \n")
		go h.hub.Post(peerevent.EV_VPConnect{
			Hostname: id.Hostname,
			Namespace:id.Namespace,
			ID:int(id.Id),
		})

	} else if !id.IsOriginal && !id.IsVP{
		//if is nvp h.hub.Post(peerevent.EV_NVPConnect{})
		h.logger.Info("post ev NVPCONNECT")
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
