package network

import (
	"crypto"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"hyperchain/common"
)

type Sec struct {
	enableTls             bool
	tlsCA                 string
	tlsServerHostOverride string
	tlsCert               string
	tlsCertPriv           string

	//client
	clientPriv *crypto.PrivateKey

	//server
	serverPriv *crypto.PrivateKey
}

//NewSec return a new sec options
func NewSec(config *viper.Viper) (*Sec, error) {
	enableTLS := config.GetBool(common.P2P_ENABLE_TLS)
	tlsCA := config.GetString(common.P2P_TLS_CA)
	tlsServerHostOverride := config.GetString(common.P2P_TLS_SERVER_HOST_OVERRIDE)
	tlsCert := config.GetString(common.P2P_TLS_CERT)
	tlsCertPriv := config.GetString(common.P2P_TLS_CERT_PRIV)

	//check the file is exist or not
	if enableTLS && !common.FileExist(tlsCA) {
		return nil, errors.New("tlsCA file not exist")
	}
	if enableTLS && !common.FileExist(tlsCert) {
		return nil, errors.New("tlsCert file not exist")
	}
	if enableTLS && !common.FileExist(tlsCertPriv) {
		return nil, errors.New("tlsCertPriv file not exist")
	}

	sec := &Sec{
		enableTls:             enableTLS,
		tlsCA:                 tlsCA,
		tlsCert:               tlsCert,
		tlsCertPriv:           tlsCertPriv,
		tlsServerHostOverride: tlsServerHostOverride,
	}

	return sec, nil
}

/**
  tls ca get dial opts and server opts part
*/

//GetGrpcClientOpts get GrpcClient options
func (s *Sec) GetGrpcClientOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	if !s.enableTls {
		logger.Warning("disable Client TLS")
		opts = append(opts, grpc.WithInsecure())
		return opts
	}
	creds, err := credentials.NewClientTLSFromFile(s.tlsCA, s.tlsServerHostOverride)
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	logger.Info("enable client TLS")
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}

//GetGrpcServerOpts get server grpc options
func (s *Sec) GetGrpcServerOpts() []grpc.ServerOption {
	var opts []grpc.ServerOption
	if !s.enableTls {
		logger.Warning("disable Server TLS")
		return opts
	}
	creds, err := credentials.NewServerTLSFromFile(s.tlsCert, s.tlsCertPriv)
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	logger.Notice("enable server TLS")
	opts = []grpc.ServerOption{grpc.Creds(creds)}
	return opts
}
