package rpc

import (

)
import (
	"reflect"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
)

type RPCManager interface {
	RegisterAllName()
	DelRegisterName()
	AddRegisterName()
}

type RPCManagerImpl struct {
	Namespace string
	Apis      []API
	services  serviceRegistry // 这是一个map，key为hpc，value为hpc的所有方法
}

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("rpc")
}

//var Apis []API
func NewRPCManagerImpl(namespace string, apis []API) *RPCManagerImpl{
	rpcmgr := &RPCManagerImpl{
			Namespace: namespace,
			Apis: apis,
		  }

	rpcmgr.init()

	return rpcmgr

}

func (rpcmgr *RPCManagerImpl) init() {
	rpcmgr.RegisterAllName()
}

func (rpcmgr *RPCManagerImpl) RegisterAllName() error{
	if rpcmgr.services == nil {
		rpcmgr.services = make(serviceRegistry)
	}

	apis := rpcmgr.Apis

	for _, api := range apis {
		if err := rpcmgr.registerName(api.Srvname, api.Service); err != nil {
			log.Errorf("registerName error: %v ", err)
			return err
		}
	}

	return nil
}

func (rpcmgr *RPCManagerImpl) registerName(name string, rcvr interface{}) error{
	svc := new(service)
	svc.typ = reflect.TypeOf(rcvr)
	rcvrVal := reflect.ValueOf(rcvr)

	if name == "" {
		return fmt.Errorf("no service name for type %s", svc.typ.String())
	}
	if !isExported(reflect.Indirect(rcvrVal).Type().Name()) {
		return fmt.Errorf("%s is not exported", reflect.Indirect(rcvrVal).Type().Name())
	}

	// already a previous service register under given sname, merge methods/subscriptions
	// 如果给定name的服务已经注册
	if regsvc, present := rpcmgr.services[name]; present {
		methods, subscriptions := suitableCallbacks(rcvrVal, svc.typ)
		if len(methods) == 0 && len(subscriptions) == 0 {
			return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
		}

		for _, m := range methods {
			regsvc.callbacks[formatName(m.method.Name)] = m
		}
		for _, s := range subscriptions {
			regsvc.subscriptions[formatName(s.method.Name)] = s
		}

		return nil
	}

	svc.name = name
	svc.callbacks, svc.subscriptions = suitableCallbacks(rcvrVal, svc.typ) // 这里的callbacks就是api.go中的方法集合

	if len(svc.callbacks) == 0 && len(svc.subscriptions) == 0 {
		return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
	}

	rpcmgr.services[svc.name] = svc

	return nil
}

func (rpcmgr *RPCManagerImpl) readRequset(RPCRequest common.RPCRequest) common.RPCResponse{
	// todo 对 request 中的参数校验

	// todo 若校验通过，则调用 Process
	return nil
}

// todo 相当于 handle()
func (rpcmgr *RPCManagerImpl) Process(RPCRequest common.RPCRequest) common.RPCResponse{
	rpcmgr.readRequset(RPCRequest)
	return nil
}
