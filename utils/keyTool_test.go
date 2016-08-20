package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestGenKeypair(t *testing.T) {
	GenKeypair()
}

func TestGetAccount(t *testing.T) {
	accounts,_ := GetAccount()
	for _,acc := range accounts{
		if assert.IsType(t,Account{},acc){
			fmt.Println("通过")
		}else{
			t.Errorf("测试失败\n")
		}
	}
}