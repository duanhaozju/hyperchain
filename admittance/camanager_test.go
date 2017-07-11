package admittance

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"crypto/ecdsa"
	"hyperchain/common"
	"hyperchain/core/crypto/primitives"
	"os"
	"strings"
	"github.com/spf13/viper"
	"fmt"
)

var (
	testCaManager *CAManager
	projectPath string
)
func init(){
	env := os.Getenv("GOPATH")
	if strings.Contains(env,":"){
		env = strings.Split(env,":")[0]
	}
	projectPath = env+"/src/hyperchain"

	fakeviper := viper.New()
	fakeviper.SetConfigFile(projectPath+"/admittance/test/global.yaml");
	fakeviper.ReadInConfig()
	fakeviper.Set("global.configs.caconfig",projectPath+"/admittance/test/caconfig.toml")

	var cmerr error
	testCaManager,cmerr = NewCAManager(fakeviper)
	if cmerr != nil {
		fmt.Println(cmerr)
		panic("cannot initliazied the camanager")
	}
}


func TestCAManager_GetECACertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECACertByte(), getFileBytes(projectPath + "/admittance/test/cert/eca.ca"))
}

func TestCAManager_GetECertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECertByte(), getFileBytes(projectPath + "/admittance/test/cert/ecert.cert"))
}

func TestCAManager_GetECertPrivateKeyByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECertPrivateKeyByte(), getFileBytes(projectPath + "/admittance/test/cert/ecert.priv"))
}

func TestCAManager_GetECertPrivKey(t *testing.T) {
	assert.IsType(t, &ecdsa.PrivateKey{}, testCaManager.GetECertPrivKey())
}

func TestCAManager_IsCheckSign(t *testing.T) {
	assert.True(t,testCaManager.IsCheckSign())
}

func TestCAManager_GetRCAcertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetRCAcertByte(), getFileBytes(projectPath + "/admittance/test/cert/rca.ca"))
}

func TestCAManager_GetIsCheckTCert(t *testing.T) {
	assert.True(t, testCaManager.IsCheckTCert())
}

func TestCAManager_GetRCertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetRCertByte(), getFileBytes(projectPath + "/admittance/test/cert/rcert.cert"))
}

//SignTCert will generate a new tcert form the camanager
func TestCAManager_SignTCert(t *testing.T) {
	publicBytes := getFileBytes(projectPath + "/admittance/test/keys/test.pub")
	signedTCert, err := testCaManager.GenTCert(common.TransportEncode(string(publicBytes)))
	assert.Nil(t, err)
	verifyResult, err2 := testCaManager.VerifyTCert(signedTCert)
	assert.Nil(t, err2)
	assert.True(t, verifyResult)

}

func TestCAManager_VerifyTCert(t *testing.T) {
	tcertBytes := getFileBytes(projectPath + "/admittance/test/cert/tcerts/tcert.cert")
	result, err := testCaManager.VerifyTCert(string(tcertBytes))
	assert.True(t, result)
	assert.Nil(t, err)
}

func TestCAManager_VerifyRCert(t *testing.T) {
	rcertBytes := getFileBytes(projectPath + "/admittance/test/cert/rcert.cert")
	result, err := testCaManager.VerifyRCert(string(rcertBytes))
	assert.True(t, result)
	assert.Nil(t, err)
}

func TestCAManager_VerifyCertSignature(t *testing.T) {
	sign, err := primitives.ECDSASign(testCaManager.GetECertPrivKey(), []byte("hyperchain"))
	assert.Nil(t, err)
	result, err := testCaManager.VerifyCertSign(string(testCaManager.GetECertByte()), []byte("hyperchain"), sign)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestCAManager_GetGrpcClientOpts(t *testing.T) {
	options := testCaManager.GetGrpcClientOpts()
	assert.NotEmpty(t, options)
}

func TestCAManager_GetGrpcServerOpts(t *testing.T) {
	options := testCaManager.GetGrpcServerOpts()
	assert.NotEmpty(t, options)
}

func TestCAManager_VerifyECert(t *testing.T) {
	ecert, err := ioutil.ReadFile("../config/cert/ecert.cert")
	if err != nil {
		t.Error("read ecert failed")
	}
	verEcert, verErr := testCaManager.VerifyECert(string(ecert))
	if verErr != nil {
		t.Error(verErr)
	}
	assert.Equal(t, true, verEcert, "verify the ecert ")
}

func getFileBytes(path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {

		panic(fmt.Sprintf("read file failed, err: %v",err));
	}
	return bytes
}