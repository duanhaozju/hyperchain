package network

import (
	"github.com/terasum/viper"
	"hyperchain/common"
	"github.com/looplab/fsm"
	"fmt"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"github.com/oleiade/lane"
	"time"
	"hyperchain/p2p/msg"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/utils"
	"strings"
	"strconv"
	"hyperchain/p2p/network/inneraddr"
)


type HyperNet struct {
	conf          *viper.Viper
	dns           *DNSResolver
	server        *Server
	clients       cmap.ConcurrentMap
	hostClientMap cmap.ConcurrentMap
	stateMachine  *fsm.FSM

	// failed queue
	failedQueue   *lane.Queue
	//reverse queue
	reverseQueue  chan [2]string

	listenPort string
	sec *Sec

	// self belong domain
	domain string

	addr *inneraddr.InnerAddr

}

func NewHyperNet(config *viper.Viper) (*HyperNet,error){
	logger = common.GetLogger(common.DEFAULT_LOG,"hypernet")
	if config == nil{
		return nil,errors.New("Readin host config failed, the viper instance is nil")
	}

	hostconf := config.GetString("global.p2p.hosts")
	port_i := config.GetInt("global.p2p.port")
	if port_i == 0{
		return nil,errors.New("invalid grpc server port")
	}
	port:= ":" + strconv.Itoa(port_i)
	if !common.FileExist(hostconf){
		fmt.Errorf("hosts config file not exist: %s",hostconf)
		return nil,errors.New(fmt.Sprintf("connot find the hosts config file: %s",hostconf))
	}
	addrconf := config.GetString("global.p2p.addr")
	if !common.FileExist(addrconf){
		fmt.Errorf("addr config file not exist: %s",hostconf)
		return nil,errors.New(fmt.Sprintf("connot find the addr config file: %s",addrconf))
	}
	ia ,domain,err := inneraddr.GetInnerAddr(addrconf)
	if err != nil {
		return nil,err
	}

	dns,err := NewDNSResolver(hostconf)
	if err != nil {
		return nil,err
	}

	sec,err := NewSec(config)
	if err !=nil{
		return nil,err
	}

	rq := make(chan [2]string)
	net :=  &HyperNet{
		dns:dns,
		server:NewServer("hypernet",rq,sec),
		hostClientMap:cmap.New(),
		failedQueue:lane.NewQueue(),
		reverseQueue:rq,
		conf:config,
		listenPort:port,
		sec:sec,
		addr:ia,
		domain:domain,
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

//Register server msg handler
//ensure this before than init server
func (hn *HyperNet)RegisterHandler(filed string,msgType pb.MsgType,handler msg.MsgHandler)error{
	return hn.server.RegisterSlot(filed,msgType,handler)
}

func (hn *HyperNet)DeRegisterHandlers(filed string){
	hn.server.DeregisterSlots(filed)
}

func (hn *HyperNet)Command(args []string,ret *[]string)error{
	if len(args) < 1{
		*ret = append(*ret,"please specific the network subcommand.")
		return nil
	}

	switch args[0] {
	case "list":{
		*ret = append(*ret,"list all connections\n")
		*ret = append(*ret,hn.dns.ListHosts()...)
	}
	case "connect":{
		if len(args)<3{
			*ret = append(*ret,"invalid connect parameters, format is `network connect [hostname] [ip:port]`",)
			break
		}
		hostname := args[1]
		ipaddr := args[2]
		if !strings.Contains(ipaddr,":"){
			*ret = append(*ret,fmt.Sprintf("%s is not a valid ipaddress, format is ipaddr:port",ipaddr))
			break
		}
		ip := strings.Split(ipaddr,":")[0]

		if !utils.IPcheck(ip){
			*ret = append(*ret,fmt.Sprintf("%s is not a valid ipv4 address",ip))
			break
		}

		port_s := strings.Split(ipaddr,":")[1]
		port,err := strconv.Atoi(port_s)
		if err != nil{
			*ret = append(*ret,fmt.Sprintf("%s valid port",port_s))
			break
		}
		*ret = append(*ret,fmt.Sprintf("connect to a new host: %s =>> %s:%d\n",hostname,ip,port))
		// real connection part
		//add dns item
		err = hn.dns.AddItem(hostname,ipaddr)
		if err != nil{
			*ret = append(*ret,fmt.Sprintf("connect to %s failed, reason: %s",hostname,err.Error()))
			break
		}
		// important after add a dns item, it should persist
		err = hn.dns.Persisit()
		if err != nil{
			*ret = append(*ret,fmt.Sprintf("filed to persist hosts file, reason: %s",err.Error()))
			break
		}else{
			*ret = append(*ret,fmt.Sprintf(" success to persist hosts file! for host %s \n",hostname))
		}

		err = hn.Connect(hostname)
		if err != nil{
			*ret = append(*ret,fmt.Sprintf("connect to %s failed, reason: %s",hostname,err.Error()))
			break
		}

		*ret = append(*ret,fmt.Sprintf("connect to %s successful.",hostname))

	}
	case "close":{
		if len(args)<2{
			*ret = append(*ret,"invalid connect parameters, format is `network close [hostname]`",)
			break
		}
		hostname := args[1]
		*ret = append(*ret,fmt.Sprintf("now closing the %s 's connection\n" ,hostname))

		err := hn.DisConnect(hostname)

		if err !=nil{
			*ret = append(*ret,fmt.Sprintf("closing the %s 's connection failed. reason: %s" ,hostname,err.Error()))
			break
		}
		err = hn.dns.Persisit()

		if err !=nil{
			*ret = append(*ret,"persist %s 's connection failed, reason: %s" ,hostname,err.Error())
			break
		}

		*ret = append(*ret,"close the host connection successed.")
	}
	case "reconnect":{
		*ret = append(*ret,"reconnect to new host")
	}
	default:
		*ret = append(*ret,fmt.Sprintf("unsupport subcommand `network %s`", args[0]))
	}
	return nil
}

//InitServer start self hypernet server listening server
func (hn *HyperNet)InitServer()error{
	hn.reverse()
	return hn.server.StartServer(hn.listenPort)
}

func (hn *HyperNet)InitClients()error{
	for _,hostname := range hn.dns.listHostnames(){
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
	if td == 0 * time.Second{
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

// if a connection failed, here will retry to connect the host name
func (hn *HyperNet)reverse() error{
	logger.Info("start reverse process")
	go func(h *HyperNet) {
		for m := range h.reverseQueue{
			hostname := m[0]
			addr := m[1]
			if _,ok := h.hostClientMap.Get(hostname);ok{
				continue
			}
			fmt.Printf("reverse connect to hostname %s,addr %s \n",hostname,addr)
			ia,err := inneraddr.InnerAddrUnSerialize([]byte(addr))
			if err != nil{
				logger.Error("cannot unserialize remote addr.")
				continue
			}
			ipaddr := ia.Get(hn.domain)

			err = h.ConnectByAddr(hostname,ipaddr)
			fmt.Println("actually connect to ",ipaddr)
			if err !=nil{
				logger.Errorf("there are something wrong when connect to host: %s",hostname)
				// TODO here should check the retry time duration, maybe is a nil
				logger.Info("It will retry connect to host",hostname,"after ",hn.conf.GetDuration("global.p2p.retrytime"))
			}else{
				logger.Info("success reverse connect to host",hostname)
			}
			err = h.dns.AddItem(hostname,ipaddr)
			if err != nil{
				logger.Errorf("cannot add a dns item into dns file %s, reason %s",hostname,err.Error())
			}
			// here when new node add should persist the connection
			err = h.dns.Persisit()
			if err != nil{
				logger.Errorf("cannot persist dns item reason %s",err.Error())
			}
		}
	}(hn)
	return nil
}

//Connect to specific host endpoint
func (hn *HyperNet)ConnectByAddr(hostname,addr string) error{
	client,err  := NewClient(addr,hn.sec)
	if err != nil{
		return err
	}
	if oldClient,ok := hn.hostClientMap.Get(hostname); ok{
		oldClient.(*Client).Close()
		hn.hostClientMap.Remove(hostname)
	}
	//err = client.Connect(nil)
	//if err != nil{
	//	return err
	//}
	hn.hostClientMap.Set(hostname,client)
	logger.Infof("success connect to %s \n",hostname)
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
	client,err  := NewClient(addr,hn.sec)
	if err != nil{
		return err
	}
	if oldClient,ok := hn.hostClientMap.Get(hostname); ok{
		oldClient.(*Client).Close()
		hn.hostClientMap.Remove(hostname)
	}
	//err = client.Connect(nil)
	//if err != nil{
	//	return err
	//}
	hn.hostClientMap.Set(hostname,client)
	logger.Infof("success connect to %s \n",hostname)
	return nil
}

//Disconnect to specific endpoint and delete the client from map
//TODO here should also handle the filed queue, find the specific host,
//TODO and cancel the retry process of this hostname
func (hn *HyperNet)DisConnect(hostname string)(err  error){
	if client, ok := hn.hostClientMap.Get(hostname);ok{
		client.(*Client).Close()
		hn.hostClientMap.Remove(hostname)
		err := hn.dns.DelItem(hostname)
		if err != nil{
			return err
		}
	}else{
		return errors.New("hostname not found.")
	}
	logger.Infof("disconnect %s successfully \n",hostname)
  	return
}

//HealthCheck check the connection is available or not at regular intervals
func (hyperNet *HyperNet)HealthCheck(){
	// TODO NetWork Health check
}

func (hypernet *HyperNet)Chat(hostname string,msg *pb.Message)error{
	hypernet.msgWrapper(msg)
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
		client.(*Client).MsgChan <- msg
		return nil
	}else{
		logger.Info("this host han't been connected. (%s)",hostname)
	}
	return errors.New("the host hasn't been initialized.")
}

func (hypernet *HyperNet)Greeting(hostname string,msg *pb.Message)(*pb.Message,error){
	hypernet.msgWrapper(msg)
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
		return client.(*Client).Greeting(msg)
	}else{
		logger.Info("this host han't been connected. (%s)",hostname)
	}
	return nil,errors.New("the host hasn't been initialized.")

}

func (hypernet *HyperNet)Whisper(hostname string,msg *pb.Message)(*pb.Message,error){
	hypernet.msgWrapper(msg)
	if client,ok := hypernet.hostClientMap.Get(hostname);ok{
			return client.(*Client).Whisper(msg)
	}else{
		logger.Info("this host han't been connected. %s",hostname)
	}
	return nil,errors.New("the host hasn't been initialized.")
}

func(hypernet *HyperNet)Stop(){
	for item := range hypernet.clients.IterBuffered() {
		client := item.Val.(*Client)
		client.Close()
		logger.Info("close client: ",item.Key)
	}
	hypernet.server.server.GracefulStop()
}


func (hn *HyperNet)msgWrapper(msg *pb.Message){
	if msg.From == nil{
		msg.From = new(pb.Endpoint)
	}
	if msg.From.Extend == nil{
		msg.From.Extend = new(pb.Extend)
	}
	if ipaddr ,err := hn.addr.Serialize();err != nil{

		msg.From.Extend.IP = []byte("{\"default\":" + "\"" + utils.GetLocalIP() + hn.listenPort + "\"}")
	}else{
		msg.From.Extend.IP = ipaddr
	}

}
