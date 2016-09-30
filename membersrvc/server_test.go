package membersrvc

import (
	"testing"
	"fmt"

	"io/ioutil"
	//"time"
	/*"io/ioutil"
	"google.golang.org/grpc/credentials"*/
	"google.golang.org/grpc/credentials"
)


func TestReadCertFile(t *testing.T) {
	/*config = loadConfig("./")
	fmt.Println(config.GetString("server.tls.cert.file"))
	creds, err := credentials.NewClientTLSFromFile(config.GetString("server.tls.cert.file"), "tlsca")
	certPEMBlock, err := ioutil.ReadFile(config.GetString("server.tls.cert.file"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(certPEMBlock)
	fmt.Println("cred",creds)*/

	//StartCAServer(config.GetString("server.tls.cert.file"),config.GetString("server.tls.key.file"))
	//StartCAServer()
	caConfig = loadConfig("./")

	fmt.Println(GetGrpcClientOpts())
	fmt.Println(GetGrpcServerOpts())



}

func TestStartCAServer(t *testing.T) {
	caConfig = loadConfig("./")
	fmt.Println(caConfig.GetString("server.tls.cert.file"))
	creds, err := credentials.NewClientTLSFromFile(caConfig.GetString("server.tls.cert.file"), "tlsca")
	certPEMBlock, err := ioutil.ReadFile(caConfig.GetString("server.tls.cert.file"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(certPEMBlock)
	fmt.Println("cred",creds)

	//StartCAServer(config.GetString("server.tls.cert.file"),config.GetString("server.tls.key.file"))
	caConfig = loadConfig("./")
	StartCAServer()



}
/*func loadConfig(path string) (config *viper.Viper) {

	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(configPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.SetConfigName("membersrvc")
	config.SetConfigType("yaml")
	//config.SetConfigName("config")
	config.AddConfigPath(path)

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", configPrefix, err))
	}

	return
}*/
