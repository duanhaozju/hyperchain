package hpc

import (
	"hyperchain/admittance"
	"hyperchain/common"
)

type CertArgs struct {
	Pubkey string `json:"pubkey"`
}

type PublicCertAPI struct {
	cm *admittance.CAManager
}

type TCertReturn struct {
	TCert string `json:"tcert"`
}

func NewPublicCertAPI(cm *admittance.CAManager) *PublicCertAPI {
	return &PublicCertAPI{
		cm: cm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicCertAPI) GetTCert(args CertArgs) (*TCertReturn, error) {
	if node.cm == nil {
		return nil, &CertError{"CAManager is nil"}
	}
	tcert, err := node.cm.SignTCert(args.Pubkey)

	if err != nil {
		log.Error("sign tcert failed")
		log.Error(err)
		return nil, &CertError{"signed tcert failed"}
	}
	tcert = common.TransportEncode(tcert)
	return &TCertReturn{TCert:tcert }, nil

}
