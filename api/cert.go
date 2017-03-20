package hpc

import (
	"hyperchain/admittance"
	"regexp"
	"hyperchain/common"
)

type CertArgs struct {
	Pubkey string `json:"pubkey"`
}

type Cert struct {
	cm *admittance.CAManager
}

type TCertReturn struct {
	TCert string `json:"tcert"`
}

func NewCertAPI(cm *admittance.CAManager) *Cert {
	return &Cert{
		cm: cm,
	}
}

// GetNodes returns status of all the nodes
func (node *Cert) GetTCert(args CertArgs) (*TCertReturn, error) {
	if node.cm == nil {
		return nil, &common.CallbackError{Message:"CAManager is nil"}
	}

	reg := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	if !reg.MatchString(args.Pubkey) {
		return nil, &common.InvalidParamsError{Message:"Invalid params, please use hex string"}
	}

	tcert, err := node.cm.SignTCert(args.Pubkey)

	if err != nil {
		log.Error(err)
		return nil, &common.CertError{Message:"Signed tcert failed"}
	}
	tcert = common.TransportEncode(tcert)
	return &TCertReturn{TCert:tcert }, nil

}