package main

import (
	"fmt"
	"github.com/terasum/viper"
)

func main() {
	vip := viper.New()
	vip.SetConfigFile("/Users/chenquan/Workspace/go/src/hyperchain/configuration/peerconfigs/peerconfig_1.toml")
	err := vip.ReadInConfig()
	if err != nil {
		panic(err)
	}

	m := vip.Get("nodes")
	m1, ok := m.([]interface{})
	fmt.Println(m1, ok)
	for _, m2 := range m1 {
		fmt.Println(m2)
		m3, ok := m2.(map[string]interface{})
		fmt.Println(m3, ok)

	}

	s := make([]interface{}, 0)
	for i := 5; i < 7; i++ {
		tmpm := make(map[string]interface{})
		tmpm["id"] = i
		tmpm["hostname"] = "example"
		s = append(s, tmpm)
	}
	fmt.Println(vip.AllKeys())
	vip.Set("nodes", s)
	fmt.Println(vip.AllKeys())
	err = vip.WriteConfig()
	if err != nil {
		panic(err)
	}
}
