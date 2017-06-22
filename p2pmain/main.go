package main

import (
	"hyperchain/p2p"
	"hyperchain/p2p/utils"
	"hyperchain/manager/event"
	"fmt"
	"os"
	"time"
	"github.com/spf13/viper"
	"strconv"
)

func main(){
	vip := viper.New()
	vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
	vip.Set("global.p2p.retrytime", "3s")
	vip.Set("global.p2p.port",50019)

	p2pManager,err := p2p.GetP2PManager(vip)
	if err != nil{
		panic("failed")
	}

	ev := new(event.TypeMux)

	peerConf := viper.New()
	peerConf.SetConfigFile(utils.GetProjectPath()+"/p2p/test/peerconfig.yaml")
	err2 := peerConf.ReadInConfig()
	if err2 != nil{
		panic(err2)
	}

	<- time.After(3*time.Second)
	peermanager,err := p2pManager.GetPeerManager("test",peerConf,ev)
	if err != nil{
		fmt.Println("err",err)
		os.Exit(1)
	}
	peermanager.Start()
	fmt.Printf("local node id is %d \n",peermanager.GetNodeId())
	fmt.Printf("local node hash is %s \n",peermanager.GetLocalNodeHash())

	peermanager.Broadcast([]byte("hello this is a test message"))

	peerlist := []uint64{1,2,3}

	peermanager.SendMsg([]byte("this unicast message"),peerlist)

	<- time.After(4*time.Second)
	t1 := time.Now()
	for i := 0;i <100000 ;i++ {
		go func(){
			now := strconv.FormatInt(time.Now().UnixNano(),10)
			times := strconv.Itoa(i)
			peermanager.SendMsg([]byte("%"+ "#" + now + "#" + times),[]uint64{1})
		}()
	}
	t2 := time.Now().UnixNano() - t1.UnixNano()

	fmt.Println("time used",t2, "ns, average: ",t2/100000)

	<- time.After(10*time.Second)
}
