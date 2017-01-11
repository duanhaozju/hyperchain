package membersrvc

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
)

var caManager *CAManager

func init(){
	var cmerr error
	caManager,cmerr = NewCAManager("../config/cert/eca.ca","../config/cert/ecert.cert","../config/cert/rcert.cert","../config/cert/rca.ca","../config/cert/ecert.priv")
	if cmerr != nil{
		panic("cannot initliazied the camanager")
	}
}

func TestCAManager_VerifyECert(t *testing.T) {
	ecert,err := ioutil.ReadFile("../config/cert/ecert.cert")
	if err != nil {
		t.Error("read ecert failed")
	}
	verEcert,verErr := caManager.VerifyECert(string(ecert))
	if verErr != nil {
		t.Error(verErr)
	}
	assert.Equal(t,true,verEcert,"verify the ecert ")
}
