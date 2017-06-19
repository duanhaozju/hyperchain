package conf

import (
	"github.com/theherk/viper"
	"fmt"
)

func ReadRouters(){
	conf := viper.New()
	conf.SetConfigFile("./peerconfig.yaml")
	conf.ReadInConfig()
	fmt.Println(conf.Get("nodes"))
}
