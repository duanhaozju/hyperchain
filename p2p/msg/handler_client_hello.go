package msg

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/csprng"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts"
	"github.com/hyperchain/hyperchain/p2p/info"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/payloads"
	"github.com/hyperchain/hyperchain/p2p/peerevent"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

type ClientHelloMsgHandler struct {
	shts     *hts.ServerHTS
	logger   *logging.Logger
	hub      *event.TypeMux
	local    *info.Info
	isOrigin bool
}

func NewClientHelloHandler(shts *hts.ServerHTS, mgrhub *event.TypeMux, localInfo *info.Info, isorigin bool, logger *logging.Logger) *ClientHelloMsgHandler {
	return &ClientHelloMsgHandler{
		shts:     shts,
		logger:   logger,
		hub:      mgrhub,
		local:    localInfo,
		isOrigin: isorigin,
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
	h.logger.Debugf("got a client hello message msgtype %s", msg.MessageType)
	h.logger.Noticef("client [%s](%s) say hello", string(msg.From.Field), string(msg.From.Hostname))
	// SERVER HELLO response
	data := []byte("hyperchain")
	esign, err := h.shts.CG.ESign(data)
	if err != nil {
		return nil, err
	}
	rsign, err := h.shts.CG.RSign(data)
	if err != nil {
		return nil, err
	}
	serverRand, err := csprng.CSPRNG(32)
	if err != nil {
		return nil, err
	}
	certpayload, err := payloads.NewCertificate(data, h.shts.CG.GetECert(), esign, h.shts.CG.GetRCert(), rsign, serverRand)
	if err != nil {
		return nil, err
	}
	// peer should
	identify := payloads.NewIdentify(h.local.IsVP, h.isOrigin, false, h.local.GetNameSpace(), h.local.Hostname, h.local.Id, certpayload)
	payload, err := identify.Serialize()
	if err != nil {
		return nil, err
	}

	rsp := pb.NewMsg(pb.MsgType_SERVERHELLO, payload)

	// NEGOTIATE shared key
	//got a identity payload
	id, err := payloads.IdentifyUnSerialize(msg.Payload)
	if err != nil {
		h.logger.Warningf("server hello error: %s \n", err.Error())
		return nil, err
	}
	h.logger.Debugf("got a remote id is rec: %v", id.IsReconnect)
	// Key Agree
	if id.Payload == nil {
		h.logger.Error("server hello id.Payload is nil")
		return nil, errors.New("server hello id.Payload is nil")
	}
	cert, err := payloads.CertificateUnMarshal(id.Payload)
	//verify
	b, e := h.shts.CG.EVerify(cert.ECert, cert.ECertSig, cert.WithData)
	if !b {
		h.logger.Critical("Verify the Ecert Signature faild", e, string(msg.From.Hostname))
		return nil, errors.New("ECert Verify failed.")
	} else {
		h.logger.Debug("Ecert Verify passed", string(msg.From.Hostname))
	}
	if id.IsVP {
		b, e = h.shts.CG.RVerify(cert.RCert, cert.RCertSig, cert.WithData)
		if !b {
			h.logger.Warningf("Verify the Rcert Signature faild", e, string(msg.From.Hostname))
			return nil, errors.New("RCert Verify failed.")
		} else {
			h.logger.Debug("RCERT Verify passed", string(msg.From.Hostname))
		}
	}
	//server rand + client rand
	r := append(serverRand, cert.Rand...)

	err = h.shts.KeyExchange(id.Hash, r, cert.ECert)
	if err != nil {
		h.logger.Errorf("handler_client_hello.go 70 %s \n", err.Error())
		return nil, errors.New(fmt.Sprintf("cannot complete the key exchange, reason: %s", err.Error()))
	}
	h.logger.Debugf(`
Server nego Key:
Local hostname: %s
Local hash %s
Peer hostname %s
Peer hash %s
server Rand %s
client Rand %s
Total rand %s
Shared key %s`, h.local.Hostname, h.local.Hash, msg.From.Hostname, id.Hash, common.ToHex(serverRand), common.ToHex(cert.Rand), common.ToHex(r), common.ToHex(h.shts.GetSK(id.Hash)))
	h.logger.Debug("isreconnect", id.IsReconnect)
	h.logger.Debug("isvp", id.IsVP)
	if id.IsVP {
		//if verify passed, should notify peer manager to reverse connect to client.
		h.logger.Debugf("post ev VPCONNECT %s rec %v", id.Hostname, id.IsReconnect)
		if id.IsReconnect {
			// if VP/NVP both should reverse to connect.
			h.logger.Infof("post ev VPCONNECT %s", id.Hostname)
			go h.hub.Post(peerevent.S_VPConnect{
				Hostname:    id.Hostname,
				Namespace:   id.Namespace,
				ID:          int(id.Id),
				IsReconnect: id.IsReconnect,
			})
		}
	} else {
		//if is nvp h.hub.Post(peerevent.EV_NVPConnect{})
		h.logger.Info("post ev NVPCONNECT", id.Hostname)
		go h.hub.Post(peerevent.S_NVPConnect{
			Hostname:    id.Hostname,
			Namespace:   id.Namespace,
			Hash:        id.Hash,
			IsReconnect: id.IsReconnect,
		})
	}

	return rsp, nil
}
