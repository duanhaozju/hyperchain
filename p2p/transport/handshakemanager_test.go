package transport

import (
	"testing"
	"os"
	"strings"
	"hyperchain/admittance"
	"github.com/terasum/viper"
	"github.com/stretchr/testify/assert"
)

var (
	testCaManager *admittance.CAManager
	projectPath string
	nHSM *HandShakeManagerNew
	remoteNHSM *HandShakeManagerNew
)
func init(){
	env := os.Getenv("GOPATH")
	if strings.Contains(env,":"){
		env = strings.Split(env,":")[0]
	}
	projectPath = env+"/src/hyperchain"

	fakeviper := viper.New()
	fakeviper.Set("global.configs.caconfig",projectPath+"/p2p/transport/test/caconfig.toml")

	var cmerr error
	testCaManager,cmerr = admittance.GetCaManager(fakeviper)
	if cmerr != nil {
		panic(cmerr)
	}
	nHSM = NewHandShakeMangerNew(testCaManager)
	remoteNHSM = NewHandShakeMangerNew(testCaManager)
	err := nHSM.GenerateSecret(remoteNHSM.GetLocalPublicKey(),"12344")
	if err != nil{
		panic(err)
	}
}


func TestNewHandShakeMangerNew(t *testing.T) {
	assert.NotNil(t,nHSM.privateKey)
	assert.NotNil(t,nHSM.publicKey)
	assert.NotNil(t,nHSM.signPublickey)
	assert.NotNil(t,nHSM.secrets)
	assert.NotNil(t,nHSM.e)
	assert.NotNil(t,nHSM.isVerified)
}


func TestHandShakeManagerNew_GenerateSecret(t *testing.T) {
	err := nHSM.GenerateSecret(remoteNHSM.GetLocalPublicKey(),"123")
	assert.Nil(t,err)
}


func TestHandShakeManagerNew_EncWithSecret(t *testing.T) {
	enced,err := nHSM.EncWithSecret([]byte("123"),"12344")
	assert.Nil(t,err)
	assert.NotNil(t,enced)
}

func TestHandShakeManagerNew_DecWithSecret(t *testing.T) {
	enced,err := nHSM.EncWithSecret([]byte("123"),"12344")
	assert.Nil(t,err)
	deced, err2 := nHSM.DecWithSecret(enced,"12344")
	assert.Nil(t,err2)
	assert.Equal(t,deced,[]byte("123"))

}

func TestHandShakeManagerNew_SetIsVerified(t *testing.T) {
	nHSM.SetIsVerified(true,"12344")
	assert.True(t,nHSM.isVerified["12344"])

}

func TestHandShakeManagerNew_GetIsVerified(t *testing.T) {
	nHSM.SetIsVerified(true,"123")
	assert.True(t,nHSM.isVerified["123"])
	assert.True(t,nHSM.GetIsVerified("12344"))
}

func TestHandShakeManagerNew_GetLocalPublicKey(t *testing.T) {
	assert.NotEmpty(t,nHSM.GetLocalPublicKey())
}

func TestHandShakeManagerNew_GetSceretPoolSize(t *testing.T) {
	assert.NotZero(t,nHSM.GetSceretPoolSize())
}

func TestHandShakeManagerNew_SetSignPublicKey(t *testing.T) {
	nHSM.SetSignPublicKey(remoteNHSM.GetLocalPublicKey(),"12344")
	assert.NotEmpty(t,nHSM.signPublickey["12344"])
}

func TestHandShakeManagerNew_GetSignPublicKey(t *testing.T) {
	nHSM.SetSignPublicKey(remoteNHSM.GetLocalPublicKey(),"12344")
	assert.NotEmpty(t,nHSM.GetSignPublicKey("12344"))

}

func TestHandShakeManagerNew_GetSecret(t *testing.T) {
	assert.NotEmpty(t,nHSM.GetSecret("12344"))
}



