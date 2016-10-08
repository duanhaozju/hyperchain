// author: Lizhong kuang
// date: 2016-09-29

package ca

import (
	"errors"

	pb "hyperchain/membersrvc/protos"
	"golang.org/x/net/context"
)

// TCAA serves the administrator GRPC interface of the TCA.
type TCAA struct {
	tca *TCA
}

// RevokeCertificate revokes a certificate from the TCA.  Not yet implemented.
func (tcaa *TCAA) RevokeCertificate(context.Context, *pb.TCertRevokeReq) (*pb.CAStatus, error) {
	Trace.Println("grpc TCAA:RevokeCertificate")

	return nil, errors.New("not yet implemented")
}

// RevokeCertificateSet revokes a certificate set from the TCA.  Not yet implemented.
func (tcaa *TCAA) RevokeCertificateSet(context.Context, *pb.TCertRevokeSetReq) (*pb.CAStatus, error) {
	Trace.Println("grpc TCAA:RevokeCertificateSet")

	return nil, errors.New("not yet implemented")
}

// PublishCRL requests the creation of a certificate revocation list from the TCA.  Not yet implemented.
func (tcaa *TCAA) PublishCRL(context.Context, *pb.TCertCRLReq) (*pb.CAStatus, error) {
	Trace.Println("grpc TCAA:CreateCRL")

	return nil, errors.New("not yet implemented")
}
