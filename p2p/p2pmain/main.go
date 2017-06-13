package main

import (
	"fmt"
	"hyperchain/p2p/network"
	"github.com/spf13/viper"
	"time"
)

func createNetwork()error{
	conf := viper.New()
	conf.SetConfigFile("./conf/global.yaml");
	err := conf.ReadInConfig()
	if err != nil{
		fmt.Errorf("readin config failed, err: %v \n",err);
		return err
	}
	hypernet,err := network.NewHyperNet(conf)
	if err != nil{
		fmt.Errorf("new hypernet failed ")
		return err
	}
	hypernet.InitServer(50012)
	err = hypernet.Connect("node1");
	if err != nil{
		return err
	}
	<- time.After(time.Second * 3)
	err = hypernet.DisConnect("host1")
	if err != nil{
		return err
	}
	<- time.After(time.Second * 3)
	err = hypernet.Connect("node3");
	if err != nil {
		return err
	}
	<- time.After(time.Second * 3)
	hypernet.DisConnect("node3")
	<- time.After(time.Second * 3)
	err = hypernet.Connect("node2");
	if err != nil {
		return err
	}
	<- time.After(time.Second * 3)
	hypernet.DisConnect("node2")
	<- time.After(time.Second * 3)
	return hypernet.Connect("node4");
}

func main(){
	fmt.Println("hello network")
	err := createNetwork()
	if err != nil{
		fmt.Printf(" create network failed, err %s \n",err.Error())
		return
	}
	fmt.Println("success")
	return
}
