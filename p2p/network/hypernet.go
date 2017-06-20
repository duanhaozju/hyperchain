package network

import (
	"github.com/spf13/viper"
	"hyperchain/common"
	"github.com/looplab/fsm"
	"fmt"
	"hyperchain/p2p/message"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"github.com/oleiade/lane"
	"time"
)

var logger = common.GetLogger(common.DEFAULT_LOG, "hypernet")

type HyperNet struct {
	conf          *viper.Viper
	dns           *DNSResolver
	server        *Server
	clients       cmap.ConcurrentMap
	hostClientMap cmap.ConcurrentMap
	stateMachine  *fsm.FSM

	// failed handler
	failedQueue   *lane.Queue
}

func NewHyperNet(config *viper.Viper) (*HyperNet,error){
	if config == nil{
		return nil,errors.New("Readin host config failed, the viper instance is nil")
	}

	hostconf := config.GetString("global.p2p.hosts")
	if !common.FileExist(hostconf){
		fmt.Errorf("pfile not exist")
		return nil,errors.New(fmt.Sprintf("connot find the hosts config file: %s",hostconf))
	}

	dns,err := NewDNSResolver(hostconf)
	if err != nil {
		return nil,err
	}
	net :=  &HyperNet{
		dns:dns,
		server:NewServer("hypernet"),
		hostClientMap:cmap.New(),
		failedQueue:lane.NewQueue(),
		conf:config,
	}
	net.stateMachine = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "Initlize", Src: []string{"created"}, Dst: "initliezed"},
			{Name: "Create", Src: []string{"established"}, Dst: "pending"},
		},
		fsm.Callbacks{
			//"enter_state":         func(e *fsm.Event) { p.enterState(e) },
			//"before_event":        func(e *fsm.Event) { p.beforeEvent(e) },
		},
	)
	err = net.retry()
	if err != nil {
		return nil,err
	}
	return net,nil
}

//InitServer start self hypernet server listening server
func (hn *HyperNet)InitServer()error{
	port := hn.conf.GetInt("global.p2p.port")
	if port == 0{
		return errors.New("invalid grpc server port")
	}
	return hn.server.StartServer(port)
}

func (hn *HyperNet)InitClients()error{
	for _,hostname := range hn.dns.ListHosts(){
		logger.Info("Now connect to host:",hostname)
		err := hn.Connect(hostname)
		if err != nil{
			logger.Error("there are something wrong when connect to host",hostname)
			// TODO here should check the retry time duration, maybe is a nil
			logger.Info("It will retry connect to host",hostname,"after ",hn.conf.GetDuration("global.p2p.retrytime"))
			hn.failedQueue.Enqueue(hostname)
		}
	}
	return nil
}

// if a connection failed, here will retry to connect the host name
func (hn *HyperNet)retry() error{
	td := hn.conf.GetDuration("global.p2p.retrytime")
	if td == 0 * time.Nanosecond{
		return errors.New("invalid time duration")
	}
	go func(h *HyperNet) {
		for range time.Tick(td){
			if h.failedQueue.Capacity() > 0{
				hostname := h.failedQueue.Dequeue().(string)
				err := h.Connect(hostname)
				if err !=nil{
					logger.Error("there are something wrong when connect to host",hostname)
					// TODO here should check the retry time duration, maybe is a nil
					logger.Info("It will retry connect to host",hostname,"after ",hn.conf.GetDuration("global.p2p.retrytime"))
					h.failedQueue.Enqueue(hostname)
				}else{
					logger.Info("success connect to host",hostname)
				}
			}
		}
	}(hn)
	return nil
}


//Connect to specific host endpoint
func (hn *HyperNet)Connect(hostname string) error{
	addr,err := hn.dns.GetDNS(hostname)
	logger.Info("connect to ",addr)
	if err != nil {
		logger.Errorf("get dns failed, err : %s \n",err.Error())
		return err
	}
	client := NewClient(addr)
	err = client.Connect()
	if err != nil{
		return err
	}
	if oldClient,ok := hn.hostClientMap.Get(hostname); ok{
		oldClient.(*Client).Close()
		hn.hostClientMap.Remove(hostname)
	}
	hn.hostClientMap.Set(hostname,client)
	logger.Infof("success connect to %s \n",hostname)
	return nil
}

//Disconnect to specific endpoint and delete the client from map
//TODO here should also handle the filed queue, find the specific host,
//TODO and cancel the retry process of this hostname
func (hn *HyperNet)DisConnect(hostname string)(err  error){
	if client, ok := hn.hostClientMap.Get(hostname);ok{
		err = client.(*Client).Close()
		hn.hostClientMap.Remove(hostname)
	}
	if err != nil {
		fmt.Printf("disconnect %s with somewrong, err: %s",hostname,err.Error())
		return
	}
	fmt.Printf("disconnect %s successfully \n",hostname)
  	return
}

//HealthCheck check the connection is available or not at regular intervals
func (hyperNet *HyperNet)HealthCheck(){
	// TODO NetWork Health check
}

func (hypernet *HyperNet)Chat(hostname string,msg *message.Message)error{
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
		client.(*Client).MsgChan <- msg
		return nil
	}
	return errors.New("the host hasn't been initialized.")
}

func (hypernet *HyperNet)Greeting(hostname string,msg *message.Message)(*message.Message,error){
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
		return client.(*Client).Greeting(msg)
	}
	return nil,errors.New("the host hasn't been initialized.")

}

func (hypernet *HyperNet)Whisper(hostname string,msg *message.Message)(*message.Message,error){
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
			return client.(*Client).Wisper(msg)
		}
		return nil,errors.New("the host hasn't been initialized.")
}
