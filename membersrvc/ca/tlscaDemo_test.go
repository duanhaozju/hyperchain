// author: Lizhong kuang
// date: 2016-09-29

package ca

import (
	"io/ioutil"

	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"google/protobuf"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"hyperchain/core/crypto/primitives"
	"hyperchain/core/util"
	membersrvc "hyperchain/membersrvc/protos"

	_ "fmt"
	"fmt"
	"strings"

)


var config1 *viper.Viper
func TestTLSDemo(t *testing.T) {


	// Skipping test for now, this is just to try tls connections
	//t.Skip()



	requestTLSCertificateDemo(t)


}

func loadConfig() ( *viper.Viper) {
	const configPrefix = "CA"
	config1 =viper.New()
	config1.SetEnvPrefix(configPrefix)
	config1.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config1.SetEnvKeyReplacer(replacer)
	config1.SetConfigName("ca")
	config1.SetConfigType("yaml")
	config1.AddConfigPath("./")

	err := config1.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error when reading %s config file: %s\n", "membersrvc", err))
	}
	return config1
}


func requestTLSCertificateDemo(t *testing.T) {
	var opts []grpc.DialOption

	config1 = loadConfig()
	fmt.Println("connect port is ",config1.GetString("peer.pki.tlsca.paddr"))

	//creds, err := credentials.NewClientTLSFromFile(viper.GetString("server.tls.cert.file"), "tlsca")
	creds, err := credentials.NewClientTLSFromFile(config1.GetString("server.tls.cert.file"), "tlsca")
	if err != nil {
		t.Logf("Failed creating credentials for TLS-CA client: %s", err)
		t.Fail()
	}

	opts = append(opts, grpc.WithTransportCredentials(creds))
	fmt.Println("connect port is ",config1.GetString("peer.pki.tlsca.paddr"))
	sockP, err := grpc.Dial(config1.GetString("peer.pki.tlsca.paddr"), opts...)
	if err != nil {
		t.Logf("Failed dialing in: %s", err)
		t.Fail()
	}

	defer sockP.Close()

	tlscaP := membersrvc.NewTLSCAPClient(sockP)

	// Prepare the request
	id := "peer"
	priv, err := primitives.NewECDSAKey()

	if err != nil {
		t.Logf("Failed generating key: %s", err)
		t.Fail()
	}

	uuid := util.GenerateUUID()

	pubraw, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	now := time.Now()
	timestamp := google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}

	req := &membersrvc.TLSCertCreateReq{
		Ts: &timestamp,
		Id: &membersrvc.Identity{Id: id + "-" + uuid},
		Pub: &membersrvc.PublicKey{
			Type: membersrvc.CryptoType_ECDSA,
			Key:  pubraw,
		}, Sig: nil}

	rawreq, _ := proto.Marshal(req)
	r, s, err := ecdsa.Sign(rand.Reader, priv, primitives.Hash(rawreq))

	if err != nil {
		t.Logf("Failed signing the request: %s", err)
		t.Fail()
	}

	R, _ := r.MarshalText()
	S, _ := s.MarshalText()
	req.Sig = &membersrvc.Signature{Type: membersrvc.CryptoType_ECDSA, R: R, S: S}

	resp, err := tlscaP.CreateCertificate(context.Background(), req)
	if err != nil {
		t.Logf("Failed requesting tls certificate: %s", err)
		t.Fail()
	}

	storePrivateKeyInClearDemo("tls_peer.priv", priv, t)
	storeCertDemo("tls_peer.cert", resp.Cert.Cert, t)
	storeCertDemo("tls_peer.ca", resp.RootCert.Cert, t)
}



func storePrivateKeyInClearDemo(alias string, privateKey interface{}, t *testing.T) {
	rawKey, err := primitives.PrivateKeyToPEM(privateKey, nil)
	if err != nil {
		t.Logf("Failed converting private key to PEM [%s]: [%s]", alias, err)
		t.Fail()
	}

	err = ioutil.WriteFile(filepath.Join(".membersrvc/", alias), rawKey, 0700)
	if err != nil {
		t.Logf("Failed storing private key [%s]: [%s]", alias, err)
		t.Fail()
	}
}

func storeCertDemo(alias string, der []byte, t *testing.T) {
	fmt.Println("entera")
	err := ioutil.WriteFile(filepath.Join(".membersrvc/", alias), primitives.DERCertToPEM(der), 0700)
	if err != nil {
		t.Logf("Failed storing certificate [%s]: [%s]", alias, err)
		t.Fail()
	}
}
