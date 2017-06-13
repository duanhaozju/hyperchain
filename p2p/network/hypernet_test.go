package network

import (
	"testing"
	"time"
	"github.com/spf13/viper"
	"fmt"
	"hyperchain/common"
)
func init(){
	conf := common.NewConfig("../test/global.yaml")
	common.InitHyperLogger(conf)
}

func TestHyperNet_InitServer(t *testing.T) {
	config := viper.New()
	config.SetConfigFile("../test/global.yaml");
	err := config.ReadInConfig()
	if err != nil{
		t.Fail()
	}
	fmt.Println(config.GetString("global.p2p.hosts"))
	hypernet,err := NewHyperNet(config)
	if err != nil{
		t.Fail()
	}
	hypernet.InitServer(50012)
	<- time.After(3* time.Second)
	t.Log("Success!")

}
