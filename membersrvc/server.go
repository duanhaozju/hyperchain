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

	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"google/protobuf"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/proto"

	"time"

	membersrvc "hyperchain/membersrvc/protos"

	"hyperchain/core/crypto"
	"hyperchain/membersrvc/ca"
	//"hyperchain/core/util"
	"hyperchain/core/crypto/primitives"

	"github.com/op/go-logging"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("membersrvc")
}

const envPrefix = "MEMBERSRVC_CA"

var caPath string
var caConfig *viper.Viper

func Start(caConfigDir string, nodeId int) {

	caConfig = LoadConfig(caConfigDir)
	caPath = caConfig.GetString("server.caserverdir")
	//if (nodeId == 1) {
	//	go StartCAServer(caConfigDir)
	//}
}

//启动ca服务器
func StartCAServer(caConfigFilepath string) {

	//viper.SetEnvPrefix(envPrefix)
	//viper.AutomaticEnv()
	//replacer := strings.NewReplacer(".", "_")
	//viper.SetEnvKeyReplacer(replacer)
	// disable the path config instead of full file path name
	// viper.SetConfigName("membersrvc")
	// viper.SetConfigType("yaml")
	// config.SetConfigName("config")
	// viper.AddConfigPath(caConfigDir)
	viper.SetConfigName("membersrvc")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.SetConfigFile(caConfigFilepath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", envPrefix, err))
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

	aca := ca.NewACA()
	defer aca.Stop()

	eca := ca.NewECA()
	defer eca.Stop()

	tca := ca.NewTCA(eca)
	defer tca.Stop()

	tlsca := ca.NewTLSCA(eca)
	defer tlsca.Stop()

	runtime.GOMAXPROCS(caConfig.GetInt("server.gomaxprocs"))

	var opts []grpc.ServerOption
	if caConfig.GetString("server.tls.cert.file") != "" {

		creds, err := credentials.NewServerTLSFromFile(caPath+caConfig.GetString("server.tls.cert.file"), caPath+caConfig.GetString("server.tls.key.file"))

		if err != nil {
			panic(err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	srv := grpc.NewServer(opts...)

	aca.Start(srv)
	eca.Start(srv)
	tca.Start(srv)
	tlsca.Start(srv)

	if sock, err := net.Listen("tcp", caConfig.GetString("server.port")); err != nil {
		ca.Error.Println("Fail to start CA Server: ", err)
		os.Exit(1)
	} else {
		srv.Serve(sock)
		sock.Close()
	}

}

//请求ca服务器端的证书
func requestTLSCertificate() {
	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile(caPath+caConfig.GetString("server.tls.cert.file"), "tlsca")
	//creds, err := credentials.NewClientTLSFromFile(clientCAPath + "/tlsca.cert", "tlsca")
	if err != nil {
		log.Info("Failed creating credentials for TLS-CA client: %s", err)

	}
	opts = append(opts, grpc.WithTransportCredentials(creds))

	sockP, err := grpc.Dial(caConfig.GetString("server.paddr"), opts...)
	if err != nil {
		fmt.Println("Failed dialing in: %s", err)

	}

	defer sockP.Close()

	tlscaP := membersrvc.NewTLSCAPClient(sockP)

	// Prepare the request
	//id := "peer"
	if err := crypto.Init(); err != nil {
		panic(fmt.Errorf("Failed initializing the crypto layer [%s]", err))
	}

	priv, err := primitives.NewECDSAKey()

	if err != nil {
		fmt.Println("Failed generating key: %s", err)

	}

	//uuid := util.GenerateUUID()

	pubraw, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	now := time.Now()
	timestamp := google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}

	req := &membersrvc.TLSCertCreateReq{
		Ts: &timestamp,
		Id: &membersrvc.Identity{Id: caConfig.GetString("node.serverhostoverride")},
		//Id: &membersrvc.Identity{Id: id + "-" + uuid},
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
	//fmt.Println("resp is", resp.Cert.Cert)

	storePrivateKeyInClear(caPath+"cert/tls_peer.priv", priv)
	storeCert(caPath+"cert/tls_peer.cert", resp.Cert.Cert)
	storeCert(caPath+"cert/tls_peer.ca", resp.RootCert.Cert)

}

//获取客户端ca配置opts
func GetGrpcClientOpts() []grpc.DialOption {

	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile(caPath+caConfig.GetString("node.tls.cap.file"), caConfig.GetString("node.serverhostoverride"))

	if err != nil {
		log.Notice("enter 222")
		log.Info("Failed creating credentials for TLS-CA client: %s", err)
		time.Sleep(time.Second * 10)
		requestTLSCertificate()
		creds, err = credentials.NewClientTLSFromFile(caPath+caConfig.GetString("node.tls.cap.file"), caConfig.GetString("node.serverhostoverride"))

		if err != nil {
			log.Fatal("can not  create credentials for TLS-CA client: %s", err)
		}

	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}

//获取服务器端ca配置opts
func GetGrpcServerOpts() []grpc.ServerOption {

	var opts []grpc.ServerOption

	creds, err := credentials.NewServerTLSFromFile(caPath+caConfig.GetString("node.tls.cert.file"), caPath+caConfig.GetString("node.tls.key.file"))

	if err != nil {
		log.Info("Failed creating credentials for TLS-CA server: %s", err)
		time.Sleep(time.Second * 10)
		requestTLSCertificate()
		creds, err = credentials.NewServerTLSFromFile(caPath+caConfig.GetString("node.tls.cert.file"), caPath+caConfig.GetString("node.tls.key.file"))

		if err != nil {
			log.Fatal("can not  create credentials for TLS-CA server: %s", err)
		}

		//panic(err)
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
func LoadConfig(caConfigFilepath string) (config *viper.Viper) {

	config = viper.New()

	// for environment variables
	//config.SetEnvPrefix(envPrefix)
	//config.AutomaticEnv()
	//replacer := strings.NewReplacer(".", "_")
	//config.SetEnvKeyReplacer(replacer)

	//config.AddConfigPath(path)
	//config.SetConfigName("membersrvc")
	//config.SetConfigType("yaml")
	//config.SetConfigName("config")
	config.SetConfigFile(caConfigFilepath)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", envPrefix, err))
	}

	return
}
