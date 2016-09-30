package membersrvc

import (
	"testing"
	//"fmt"

	"time"
	/*"io/ioutil"
	"google.golang.org/grpc/credentials"*/
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
	StartCAServer()
	time.Sleep(time.Second * 20)
	a := make(chan bool)
	requestTLSCertificateDemo()
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 5)
			requestTLSCertificateDemo()

		}
	}()
	<-a

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
