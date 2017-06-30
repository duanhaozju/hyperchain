package network

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Sec struct {
	enableTls bool
	tlsCA string
	tlsServerHostOverride string
	tlsCert string
	tlsCertPriv string
}


//NewSec return a new sec options
func NewSec(enableTls bool,tlsCA string, tlsServerHostOverride string, tlsCert string, tlsCertPriv string) (*Sec,error){
	//check node



	return &Sec{
		enableTls: enableTls,
		tlsCA:tlsCA,
		tlsCert:tlsCert,
		tlsCertPriv:tlsCertPriv,
		tlsServerHostOverride:tlsServerHostOverride,
	}
}


/**
  tls ca get dial opts and server opts part
 */

//GetGrpcClientOpts get GrpcClient options
func (s *Sec) GetGrpcClientOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	if !s.enableTls{
		logger.Warning("disable Client TLS")
		opts = append(opts,grpc.WithInsecure())
		return opts
	}
	creds, err := credentials.NewClientTLSFromFile(s.tlsCA, s.tlsServerHostOverride)
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}

//GetGrpcServerOpts get server grpc options
func (s *Sec) GetGrpcServerOpts() []grpc.ServerOption {
	var opts []grpc.ServerOption
	if !s.enableTls{
		logger.Warning("disable Server TLS")
		return opts
	}
	creds, err := credentials.NewServerTLSFromFile(s.tlsCert, s.tlsCertPriv)
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	opts = []grpc.ServerOption{grpc.Creds(creds)}
	return opts
}

