package network

import (
	"github.com/spf13/viper"
	"hyperchain/common"
	"sync"
	"github.com/looplab/fsm"
	"fmt"
)

var logger = common.GetLogger(common.DEFAULT_LOG, "hypernet")
type HyperNet struct {
	dns *DNSResolver
	server *Server
	clients map[string]*Client
	idClientMap map[string]string
	idClientMapLock sync.Mutex
	stateMachine *fsm.FSM
}

func init(){
	globalConfig := common.NewConfig("../test/global.yaml")
	common.InitHyperLoggerManager(globalConfig)
}

func NewHyperNet(config *viper.Viper) (*HyperNet,error){
	dns,err := NewDNSResolver(config.GetString("global.p2p.hosts"))
	if err != nil {
	   fmt.Println("Readin hosts config failed, error :",err)
		return nil,err
	}
	net :=  &HyperNet{
		dns:dns,
		server:NewServer(),
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
	return net,nil
}

//InitServer start self hypernet server listening server
func (hyperNet *HyperNet)InitServer(port int){
	go hyperNet.server.StartServer(port)
}

//Bind the higher layer identifier
func (hyperNet *HyperNet)Bind(identifier string, hostname string){
	//TODO Bind should ensure client or server cert to ensure the
	//TODO node is legal
	hyperNet.idClientMapLock.Lock()
	hyperNet.idClientMapLock.Unlock()
	hyperNet.idClientMap[identifier] = hostname
}

func (hyperNet *HyperNet)Unbind(identifier string){
	hyperNet.idClientMapLock.Lock()
	defer hyperNet.idClientMapLock.Unlock()
	delete(hyperNet.idClientMap,identifier)
}

//Client use specitic hostname to crete a client
func(hyperNet *HyperNet)CreateClient(hostname string) (error){
	addr,err :=  hyperNet.dns.GetDNS(hostname)
	if err != nil{
		logger.Errorf("Cannot create hypernet client for hostname %s , error info: %v \n",hostname,err)
		return err
	}
	hyperNet.clients[hostname] = NewClient(addr)
	return nil
}

//HealthCheck check the connection is available or not at regular intervals
func (HyperNet *HyperNet)HealthCheck(){
	// TODO NetWork Health check
}
