package hpc

import (
	"hyperchain/membersrvc"
	"hyperchain/common"
)

type CertArgs struct {
	Pubkey string `json:"pubkey"`
}

type PublicCertAPI struct {
	cm *membersrvc.CAManager
}

type TCertReturn struct {
	TCert string `json:"tcert"`
}

func NewPublicCertAPI(cm *membersrvc.CAManager) *PublicCertAPI {
	return &PublicCertAPI{
		cm: cm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicCertAPI) GetTCert(args CertArgs) (TCertReturn, error) {
	if node.cm == nil {
		return TCertReturn{TCert: "invalid tcert"}, &CertError{"CAManager is nil"}
	}
	tcert, err := node.cm.SignTCert(args.Pubkey)

	if err != nil {
		log.Error("sign tcert failed")
		log.Error(err)
		return TCertReturn{TCert: ""}, &CertError{"signed tcert failed"}
	}
	tcert = common.EncodeUriComponent(tcert)
	return TCertReturn{TCert:tcert }, nil

}
