package SMCrypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var smc *SMCrypto

func init(){
	smc = NewSMCrypto()
}

func TestSMCrypto_Connect(t *testing.T) {
	ret := smc.Connect()
	assert.Equal(t,ret,nil)
}

func TestSMCrypto_GenKeyPair(t *testing.T) {
	err := smc.Connect()
	if err != nil{
		t.Fatal(err)
	}
	pub,pri,err := smc.GenKeyPair();
	if err != nil {
		t.Fatalf("gen failed,err %v\n",err)
	}
	t.Log("pub",pub)
	assert.Equal(t,len(pub),136)
	t.Log("pri",pri)
	assert.Equal(t,len(pri),136)
}


func TestSMCrypto_GenSessionKey(t *testing.T) {
	err := smc.Connect()
	if err != nil{
		t.Fatal(err)
	}
	pub,pri,err := smc.GenKeyPair();
	if err != nil {
		t.Fatalf("gen failed,err %v\n",err)
	}
	t.Log("pub",pub)
	assert.Equal(t,len(pub),136)
	t.Log("pri",pri)
	assert.Equal(t,len(pri),136)

	sessionKey,KCV ,err := smc.GenSessionKey(pub)
	t.Log(string(sessionKey))
	t.Log(string(KCV))
}

func BenchmarkSMCrypto_Hash(b *testing.B) {

}