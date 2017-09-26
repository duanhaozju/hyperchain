package api

import (
	"hyperchain/admittance"
	"hyperchain/common"
	"regexp"
)

/*
    This file implements the handler of certificate service API
	which can be invoked by client in JSON-RPC request.
 */

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
		return nil, &common.InvalidParamsError{Message: "Invalid params, please use hex string"}
	}

	tcert, err := node.cm.GenTCert(args.Pubkey)
	if err != nil {
		log.Error(err)
		return nil, &common.CertError{Message: "Signed tcert failed"}
	}

	err = admittance.RegisterCert([]byte(tcert))
	if err != nil {
		log.Error(err)
		return nil, &common.CertError{Message: "Register tcert failed"}
	}
	tcert = common.TransportEncode(tcert)
	return &TCertReturn{TCert: tcert}, nil

}
