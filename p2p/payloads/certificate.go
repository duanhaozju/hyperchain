package payloads

import "github.com/golang/protobuf/proto"
//release 1.3 => 13
const CERT_VERSION=int64(13)

func NewCertificate(data ,ecert, ecertsign, rcert,rcertsign []byte)([]byte,error){
	cert :=  &Certificate{
		Version:CERT_VERSION,
		ECert:ecert,
		ECertSig:ecertsign,
		RCert:rcert,
		RCertSig:rcertsign,
		WithData:data,
	}
	return proto.Marshal(cert)
}

func CertificateUnMarshal(raw []byte)(*Certificate,error){
	var cert = new(Certificate)
	err := proto.Unmarshal(raw,cert)
	if err !=nil{
		return nil,err
	}
	return cert,nil
}
