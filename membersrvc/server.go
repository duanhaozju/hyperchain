// author: Lizhong kuang
// date: 2016-09-29

package membersrvc

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

var caConfig *viper.Viper

func StartCAServer() {
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
		//cfgpath := filepath.Join(p, "src/github.com/hyperledger/fabric/membersrvc")
		cfgpath := filepath.Join(p, "src/hyperchain/membersrvc/")

		fmt.Println(cfgpath)
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
		//creds, err := credentials.NewServerTLSFromFile(cert, priv)


		creds, err := credentials.NewServerTLSFromFile(viper.GetString("server.tls.cert.file"), viper.GetString("server.tls.key.file"))

		if err != nil {
			panic(err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	srv := grpc.NewServer(opts...)
	fmt.Println("port is ", viper.GetString("server.port"))

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
	time.Sleep(time.Second * 20)
	a := make(chan bool)
	requestTLSCertificateDemo()
	/*go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 5)
			requestTLSCertificateDemo()

		}
	}()*/
	<-a

}


func requestTLSCertificateDemo() {
	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile(viper.GetString("server.tls.cert.file"), "tlsca")
	//creds, err := credentials.NewClientTLSFromFile(clientCAPath + "/tlsca.cert", "tlsca")
	if err != nil {
		fmt.Println("Failed creating credentials for TLS-CA client: %s", err)

	}

	opts = append(opts, grpc.WithTransportCredentials(creds))
	fmt.Println("connect port is ", viper.GetString("server.paddr"))
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
	fmt.Println("resp is", resp.Cert.Cert)

	storePrivateKeyInClear("cert/tls_peer.priv", priv)
	storeCert("cert/tls_peer.cert", resp.Cert.Cert)
	storeCert("cert/tls_peer.ca", resp.RootCert.Cert)

}

func GetGrpcClientOpts() []grpc.DialOption {

	var opts []grpc.DialOption

	//creds, err := credentials.NewClientTLSFromFile(viper.GetString("server.tls.cert.file"), "tlsca")

	creds, err := credentials.NewClientTLSFromFile("./membersrvc/"+caConfig.GetString("node.tls.cap.file"), "peer-18f4f227-3ab9-4531-a469-d6f174d88cf3")


	if err != nil {
		fmt.Println("Failed creating credentials for TLS-CA client: %s", err)

	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}
func Start(){
	caConfig = loadConfig("./membersrvc/")
}
func GetGrpcServerOpts() []grpc.ServerOption {

	var opts []grpc.ServerOption

	creds, err := credentials.NewServerTLSFromFile("./membersrvc/"+caConfig.GetString("node.tls.cert.file"),"./membersrvc/"+ caConfig.GetString("node.tls.key.file"))


	//creds, err := credentials.NewServerTLSFromFile(viper.GetString("server.tls.cert.file"), viper.GetString("server.tls.key.file"))

	if err != nil {
		panic(err)
	}
	opts = []grpc.ServerOption{grpc.Creds(creds)}
	return opts

}

func storePrivateKeyInClear(alias string, privateKey interface{}) {
	rawKey, err := primitives.PrivateKeyToPEM(privateKey, nil)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(filepath.Join("./", alias), rawKey, 0700)
	if err != nil {
		fmt.Println(err)
	}
}

func storeCert(alias string, der []byte) {
	fmt.Println("entera")
	err := ioutil.WriteFile(filepath.Join("./", alias), primitives.DERCertToPEM(der), 0700)
	if err != nil {
		fmt.Println(err)
	}
}
func loadConfig(path string) (config *viper.Viper) {

	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(envPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.SetConfigName("membersrvc")
	config.SetConfigType("yaml")
	//config.SetConfigName("config")
	config.AddConfigPath(path)

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", envPrefix, err))
	}

	return
}