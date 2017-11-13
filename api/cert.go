package api

import (
	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/common"
	"regexp"
)

// This file implements the handler of Certificate service API which
// can be invoked by client in JSON-RPC request.

type CertArgs struct {
	Pubkey string `json:"pubkey"`
}

type Cert struct {
	cm        *admittance.CAManager
	namespace string
}

type TCertReturn struct {
	TCert string `json:"tcert"`
}

func NewCertAPI(namespace string, cm *admittance.CAManager) *Cert {
	return &Cert{
		cm:        cm,
		namespace: namespace,
	}
}

// GetTCert creates a new tcert for the given public key.
func (node *Cert) GetTCert(args CertArgs) (*TCertReturn, error) {
	log := common.GetLogger(node.namespace, "api")
	if node.cm == nil {
		return nil, &common.CallbackError{Message: "CAManager is nil"}
	}

	reg := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	if !reg.MatchString(args.Pubkey) {
		return nil, &common.InvalidParamsError{Message: "invalid params, please use hex string"}
	}

	tcert, err := node.cm.GenTCert(args.Pubkey)
	if err != nil {
		log.Errorf("generate tcert error: %v", err)
		return nil, &common.CertError{Message: "generated tcert failed"}
	}

	err = admittance.RegisterCert([]byte(tcert))
	if err != nil {
		log.Errorf("register tcert error: %v", err)
		return nil, &common.CertError{Message: "register tcert failed"}
	}
	tcert = common.TransportEncode(tcert)
	return &TCertReturn{TCert: tcert}, nil

}
