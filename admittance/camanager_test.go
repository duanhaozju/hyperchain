package admittance

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/primitives"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	testCaManager *CAManager
	projectPath   string
)

func init() {
	env := os.Getenv("GOPATH")
	if strings.Contains(env, ":") {
		env = strings.Split(env, ":")[0]
	}
	projectPath = env + "/src/github.com/hyperchain/hyperchain"
	config := common.NewConfig(projectPath + "/admittance/test/namespace.toml")
	var cmerr error
	config.Set(common.NAMESPACE, "")
	err := hyperdb.InitDatabase(config, "global")
	if err != nil {
		fmt.Println(err)
	}
	testCaManager, cmerr = NewCAManager(config)
	if cmerr != nil {
		fmt.Println(cmerr)
		panic("cannot initliazied the camanager")
	}
}

func getFileBytes(path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {

		panic(fmt.Sprintf("read file failed, err: %v", err))
	}
	return bytes
}

func TestNewCAManager(t *testing.T) {
	env := os.Getenv("GOPATH")
	if strings.Contains(env, ":") {
		env = strings.Split(env, ":")[0]
	}
	projectPath = env + "/src/github.com/hyperchain/hyperchain"
	config := common.NewConfig(projectPath + "/admittance/test/namespace.toml")
	config.Set(common.NAMESPACE, "")
	err := hyperdb.InitDatabase(config, "global")
	if err != nil {
		fmt.Println(err)
	}
	_, cmerr := NewCAManager(config)
	assert.Nil(t, cmerr)
}

func TestCAManager_GetECACertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECACertByte(), getFileBytes(projectPath+"/admittance/test/certs/eca.cert"))
}

func TestCAManager_GetECertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECertByte(), getFileBytes(projectPath+"/admittance/test/certs/ecert.cert"))
}

func TestCAManager_GetECertPrivateKeyByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetECertPrivateKeyByte(), getFileBytes(projectPath+"/admittance/test/certs/ecert.priv"))
}

func TestCAManager_GetECertPrivKey(t *testing.T) {
	assert.IsType(t, &ecdsa.PrivateKey{}, testCaManager.GetECertPrivKey())
}

func TestCAManager_IsCheckSign(t *testing.T) {
	assert.True(t, testCaManager.IsCheckSign())
}

func TestCAManager_GetRCAcertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetRCAcertByte(), getFileBytes(projectPath+"/admittance/test/certs/rca.cert"))
}

func TestCAManager_GetIsCheckTCert(t *testing.T) {
	assert.True(t, testCaManager.IsCheckTCert())
}

func TestCAManager_GetRCertByte(t *testing.T) {
	assert.Equal(t, testCaManager.GetRCertByte(), getFileBytes(projectPath+"/admittance/test/certs/rcert.cert"))
}

//SignTCert will generate a new tcert form the camanager
func TestCAManager_SignTCert(t *testing.T) {
	publicBytes := getFileBytes(projectPath + "/admittance/test/keys/test.pub")
	_, err := testCaManager.GenTCert(common.TransportEncode(string(publicBytes)))
	assert.Nil(t, err)
}

func TestCAManager_VerifyRCert(t *testing.T) {
	rcertBytes := getFileBytes(projectPath + "/admittance/test/certs/rcert.cert")
	result, err := testCaManager.VerifyRCert(string(rcertBytes))
	assert.True(t, result)
	assert.Nil(t, err)
}

func TestCAManager_VerifyRCert2(t *testing.T) {
	result, err := testCaManager.VerifyRCert("hyperchain")
	fmt.Println(result, err)
	assert.False(t, result)
	assert.NotNil(t, err)
}

func TestCAManager_VerifyCertSignature(t *testing.T) {
	sign, err := primitives.ECDSASign(testCaManager.GetECertPrivKey(), []byte("hyperchain"))
	assert.Nil(t, err)
	result, err := testCaManager.VerifyCertSign(string(testCaManager.GetECertByte()), []byte("hyperchain"), sign)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestCAManager_VerifyECert(t *testing.T) {
	ecert, err := ioutil.ReadFile("./test/certs/ecert.cert")
	if err != nil {
		t.Error("read ecert failed")
	}
	verEcert, verErr := testCaManager.VerifyECert(string(ecert))
	if verErr != nil {
		t.Error(verErr)
	}
	assert.Equal(t, true, verEcert, "verify the ecert ")
}

func TestListDir(t *testing.T) {
	_, err := ListDir("test/certs", "cert")
	assert.Nil(t, err)
}

func TestCAManager_VerifyTCert(t *testing.T) {
	publicBytes := getFileBytes(projectPath + "/admittance/test/keys/test.pub")
	signedTCert, err := testCaManager.GenTCert(common.TransportEncode(string(publicBytes)))
	assert.Nil(t, err)
	err = RegisterCert([]byte(signedTCert))
	assert.Nil(t, err)
	verifyResult, err2 := testCaManager.VerifyTCert(signedTCert, "SendTx")
	assert.Nil(t, err2)
	assert.True(t, verifyResult)
}

func TestCAManager_VerifyTCert2(t *testing.T) {
	sdkcert, err := ioutil.ReadFile("./test/certs/sdkcert.cert")
	if err != nil {
		t.Error("read ecert failed")
	}
	verifyResult, err2 := testCaManager.VerifyTCert(string(sdkcert), "getTCert")
	assert.Nil(t, err2)
	assert.True(t, verifyResult)
}

func TestDeleteTestData(t *testing.T) {
	if common.FileExist("namespaces") {
		err := os.RemoveAll("namespaces")
		if err != nil {
			t.Fatalf("delete dir error: %v", err)
		}
	}
}
