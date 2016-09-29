/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"strings"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"google/protobuf"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"


	"time"

	//membersrvc "github.com/hyperledger/fabric/membersrvc/protos"
        membersrvc "hyperchain/membersrvc/protos"

	"hyperchain/membersrvc/ca"
	"hyperchain/core/crypto"
	"hyperchain/core/util"
	"hyperchain/core/crypto/primitives"
)

const envPrefix = "MEMBERSRVC_CA"

func main() {
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigName("membersrvc")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	// Path to look for the config file based on GOPATH
	gopath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(gopath) {
		cfgpath := filepath.Join(p, "src/github.com/hyperledger/fabric/membersrvc")
		viper.AddConfigPath(cfgpath)
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error when reading %s config file: %s\n", "membersrvc", err))
	}

	var iotrace, ioinfo, iowarning, ioerror, iopanic io.Writer
	if viper.GetInt("logging.trace") == 1 {
		iotrace = os.Stdout
	} else {
		iotrace = ioutil.Discard
	}
	if viper.GetInt("logging.info") == 1 {
		ioinfo = os.Stdout
	} else {
		ioinfo = ioutil.Discard
	}
	if viper.GetInt("logging.warning") == 1 {
		iowarning = os.Stdout
	} else {
		iowarning = ioutil.Discard
	}
	if viper.GetInt("logging.error") == 1 {
		ioerror = os.Stderr
	} else {
		ioerror = ioutil.Discard
	}
	if viper.GetInt("logging.panic") == 1 {
		iopanic = os.Stdout
	} else {
		iopanic = ioutil.Discard
	}

	// Init the crypto layer
	if err := crypto.Init(); err != nil {
		panic(fmt.Errorf("Failed initializing the crypto layer [%s]", err))
	}

	ca.LogInit(iotrace, ioinfo, iowarning, ioerror, iopanic)
	// cache configure
	ca.CacheConfiguration()

	ca.Info.Println("CA Server (" + viper.GetString("server.version") + ")")

	aca := ca.NewACA()
	defer aca.Stop()

	eca := ca.NewECA()
	defer eca.Stop()

	tca := ca.NewTCA(eca)
	defer tca.Stop()

	tlsca := ca.NewTLSCA(eca)
	defer tlsca.Stop()

	runtime.GOMAXPROCS(viper.GetInt("server.gomaxprocs"))

	var opts []grpc.ServerOption
	if viper.GetString("server.tls.cert.file") != "" {
		creds, err := credentials.NewServerTLSFromFile("/Users/kuang/go-workspace/src/github.com/hyperledger/fabric/membersrvc/ca/test_resources/tlsca.cert", "/Users/kuang/go-workspace/src/github.com/hyperledger/fabric/membersrvc/ca/test_resources/tlsca.priv")

		//creds, err := credentials.NewServerTLSFromFile(viper.GetString("server.tls.cert.file"), viper.GetString("server.tls.key.file"))
		if err != nil {
			panic(err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	srv := grpc.NewServer(opts...)
	fmt.Println("port is ",viper.GetString("server.port"))

	aca.Start(srv)
	eca.Start(srv)
	tca.Start(srv)
	tlsca.Start(srv)

	if sock, err := net.Listen("tcp", viper.GetString("server.port")); err != nil {
		ca.Error.Println("Fail to start CA Server: ", err)
		os.Exit(1)
	} else {
		go srv.Serve(sock)
		//sock.Close()
	}
	a:=make(chan bool)
	time.Sleep(time.Second * 10)

	requestTLSCertificateDemo()
	go func(){
		for i:=0;i<10;i++{
			time.Sleep(time.Second *5)
			requestTLSCertificateDemo()


		}
	}()

	<-a
}


func requestTLSCertificateDemo() {
	var opts []grpc.DialOption



	//creds, err := credentials.NewClientTLSFromFile(viper.GetString("server.tls.cert.file"), "tlsca")
	creds, err := credentials.NewClientTLSFromFile("/Users/kuang/go-workspace/src/github.com/hyperledger/fabric/membersrvc/ca/test_resources/tlsca.cert", "tlsca")
	if err != nil {
		fmt.Println("Failed creating credentials for TLS-CA client: %s", err)

	}

	opts = append(opts, grpc.WithTransportCredentials(creds))
	fmt.Println("connect port is ",viper.GetString("server.paddr"))
	sockP, err := grpc.Dial(viper.GetString("server.paddr"), opts...)
	if err != nil {
		fmt.Println("Failed dialing in: %s", err)

	}

	defer sockP.Close()

	tlscaP := membersrvc.NewTLSCAPClient(sockP)

	// Prepare the request
	id := "peer"
	priv, err := primitives.NewECDSAKey()

	if err != nil {
		fmt.Println("Failed generating key: %s", err)

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
		fmt.Println("Failed dialing in: %s", err)
	}

	R, _ := r.MarshalText()
	S, _ := s.MarshalText()
	req.Sig = &membersrvc.Signature{Type: membersrvc.CryptoType_ECDSA, R: R, S: S}

	resp, err := tlscaP.CreateCertificate(context.Background(), req)
	if err != nil {
		fmt.Println("Failed dialing in: %s", err)
	}
	fmt.Println("resp is",resp.Cert.Cert)

	storePrivateKeyInClear("tls_peer.priv", priv)
	storeCert("tls_peer.cert", resp.Cert.Cert)
	storeCert("tls_peer.ca", resp.RootCert.Cert)
}
func storePrivateKeyInClear(alias string, privateKey interface{}) {
	rawKey, err := primitives.PrivateKeyToPEM(privateKey, nil)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(filepath.Join("/membersrvc/", alias), rawKey, 0700)
	if err != nil {
		fmt.Println(err)
	}
}

func storeCert(alias string, der []byte) {
	fmt.Println("entera")
	err := ioutil.WriteFile(filepath.Join("membersrvc/", alias), primitives.DERCertToPEM(der), 0700)
	if err != nil {
		fmt.Println(err)
	}
}
