package hts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/terasum/viper"
	"hyperchain/p2p/utils"
	"testing"
)

//func initialCertGroup() {
//
//}

func TestCertGroup_ESignAndEVerify(t *testing.T) {
	nsCnfigPath := utils.GetProjectPath() + "/p2p/test/namespace2.toml"

	vip := viper.New()
	vip.SetConfigFile(nsCnfigPath)
	err := vip.ReadInConfig()
	if err != nil {
		t.Fatalf(fmt.Sprintf("cannot read in the caconfig, reason: %s ", err.Error()))
	}

	cg, err := NewCertGroup("", vip)
	if err != nil {
		t.Fatalf("cannot create a new CertGroup, error: %v", err.Error())
	}

	ecert := cg.GetECert()
	assert.NotNil(t, ecert, "should get ecert")

	rcert := cg.GetRCert()
	assert.NotNil(t, rcert, "should get rcert")

	data := []byte("hyperchain")
	sign, err := cg.ESign(data)
	if err != nil {
		t.Fatalf("ESign error: %v", err)
	}

	isSuccess, err := cg.EVerify(cg.eCERT, sign, data)
	if err != nil {
		t.Fatalf("EVerify error: %v", err)
	}
	assert.True(t, isSuccess, "should verify successfully")

	cg.sign = false
	sign, err = cg.ESign(data)
	assert.Nil(t, err, "shouldn't return error, because it needn't sign")
	assert.Nil(t, sign, "shouldn't return signature, because it needn't sign")

	isSuccess, err = cg.EVerify(cg.eCERT, sign, data)
	assert.Nil(t, err, "should n't return error, because it needn't verify")
	assert.True(t, isSuccess, "should pass verification, because it needn't verify")
}

func TestCertGroup_RSignAndRVerify(t *testing.T) {
	nsCnfigPath := utils.GetProjectPath() + "/p2p/test/namespace2.toml"

	vip := viper.New()
	vip.SetConfigFile(nsCnfigPath)
	err := vip.ReadInConfig()
	if err != nil {
		t.Fatalf(fmt.Sprintf("cannot read in the caconfig, reason: %s ", err.Error()))
	}

	cg, err := NewCertGroup("", vip)
	if err != nil {
		t.Fatalf("cannot create a new CertGroup, error: %v", err.Error())
	}

	data := []byte("hyperchain")
	sign, err := cg.RSign(data)
	if err != nil {
		t.Fatalf("RSign error: %v", err)
	}

	isSuccess, err := cg.RVerify(cg.rCERT, sign, data)
	if err != nil {
		t.Fatalf("RVerify error: %v", err)
	}
	assert.True(t, isSuccess, "should verify successfully")

	cg.sign = false
	sign, err = cg.RSign(data)
	assert.Nil(t, err, "shouldn't return error, because it needn't sign")
	assert.Nil(t, sign, "shouldn't return signature, because it needn't sign")

	isSuccess, err = cg.RVerify(cg.rCERT, sign, data)
	assert.Nil(t, err, "should n't return error, because it needn't verify")
	assert.True(t, isSuccess, "should pass verification, because it needn't verify")

}
