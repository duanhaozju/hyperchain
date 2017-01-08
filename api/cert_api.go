package hpc

import (
	"hyperchain/membersrvc"
)

type CertArgs struct {
	 pubkey string `json:"pubkey"`
}

type PublicCertAPI struct{
	cm *membersrvc.CAManager
}

type TcertReturn struct {
	TCert string `json:"tcert"`
}

func NewPublicCertAPI( cm *membersrvc.CAManager) *PublicCertAPI{
	return &PublicCertAPI{
		cm: cm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicCertAPI) GetTCert(args CertArgs) (TcertReturn, error) {
	if node.cm == nil {
		return TcertReturn{TCert:"invalid tcert"}, &CertError{"CAManager is nil"}
	}
	tcert,err :=  node.cm.SignTCert(args.pubkey)
	if err != nil{
		log.Error("sign tcert failed")
		log.Error(err)
		return "", &CertError{"signed tcert failed"}
	}

	return TcertReturn{TCert:tcert},nil

}

