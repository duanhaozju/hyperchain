package api

import (
	"hyperchain/admittance"
	"hyperchain/common"
	"regexp"
	"hyperchain/hyperdb"
	"encoding/asn1"
)

const CertKey string = "tcerts"

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

type RegisterTcerts struct {
	Tcerts []string
}

// GetNodes returns status of all the nodes
func (node *Cert) GetTCert(args CertArgs) (*TCertReturn, error) {
	log := common.GetLogger(node.namespace, "api")
	if node.cm == nil {
		return nil, &common.CallbackError{Message: "CAManager is nil"}
	}

	reg := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	if !reg.MatchString(args.Pubkey) {
		return nil, &common.InvalidParamsError{Message: "Invalid params, please use hex string"}
	}

	//tcert, err := node.cm.SignTCert(args.Pubkey)
	tcert, err := node.cm.GenTCert(args.Pubkey)
	if err != nil {
		log.Error(err)
		return nil, &common.CertError{Message: "Signed tcert failed"}
	}
	err = RegisterCert([]byte(tcert))
	if err != nil {
		log.Error(err)
		return nil, &common.CertError{Message: "Register tcert failed"}
	}
	tcert = common.TransportEncode(tcert)
	return &TCertReturn{TCert: tcert}, nil

}

func RegisterCert(tcert []byte) error {
	log := common.GetLogger(common.DEFAULT_NAMESPACE, "api")
	db,err := hyperdb.GetDBDatabase()
	if err!=nil {
		log.Error(err)
		return &common.CertError{Message: "Get Database failed"};
	}
	certs,err := db.Get([]byte(CertKey))
	tcertStr := string(tcert)
	//First to Save CertList
	if err != nil{
		//log.Critical("Register TCERT:",tcertStr)
		regLists := RegisterTcerts{[]string{tcertStr}}
		lists,err := asn1.Marshal(regLists)
		if err!= nil{
			log.Error(err)
			return &common.CertError{Message: "Marshal cert lists failed"};
		}
		err = db.Put([]byte(CertKey),lists)
		if err!= nil{
			log.Error(err)
			return &common.CertError{Message: "Save cert lists failed"};
		}
		return nil
	}
	//log.Critical("GET CERT LIST FROM DB:",certs)
	Regs := struct {
		Tcerts []string
	}{}
	_,err = asn1.Unmarshal(certs,&Regs)
	if err!=nil {
		log.Error(err)
		return  &common.CertError{Message: "UnMarshal cert lists failed"};
	}
	Regs.Tcerts = append(Regs.Tcerts,tcertStr)
	lists,err := asn1.Marshal(Regs)
	if err!= nil{
		log.Error(err)
		return &common.CertError{Message: "Marshal cert lists failed"};
	}
	err = db.Put([]byte(CertKey),lists)
	if err!= nil{
		log.Error(err)
		return &common.CertError{Message: "Save cert lists failed"};
	}
	return nil
}
