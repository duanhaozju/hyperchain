//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

// Package namespace provides mechanism to manage namespaces and request process
package namespace

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/common/service/server"
	"hyperchain/core/ledger/bloom"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"hyperchain/common/interface"
)

// This file defines the Namespace Manager interface, which managers all
// the namespaces this node has participated in. Start/stop NamespaceManager
// will start/stop all namespaces registered to Namespace Manager and then
// start/stop jvm manager which manages jvm executor. So be careful when
// system administrator wants to stop NamespaceManager.

// Nodes join into a namespace by register and then start, exit a namespace
// by stop and then de-register from that namespace. De-registration from
// a namespace directly will first stop that namespace and then de-register
// by default. So life cycle of a namespace should be like:
// register ===> start ===> running ===> stop ===> de-register.

// Requests sent from clients will be dispatched to different namespace's
// request processor with the given namespace name.

var logger *logging.Logger

func init() {
	logger = common.GetLogger(common.DEFAULT_LOG, "nsmgr")
}

const (
	DEFAULT_NAMESPACE  = "global"
	NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

var once sync.Once
var nr NamespaceManager

// NamespaceManager manages all namespaces registered to this node.
type NamespaceManager interface {
	// Start starts the namespace manager service.
	Start() error

	// Stop stops the namespace manager service.
	Stop() error

	// List returns all namespace names in system.
	List() []string

	// Register registers a newed namespace by the given name, if the namespace
	// has been registered to system before or the config path of that
	// namespace(./namespace/name) doesn't exist, returns an error, else,
	// constructs newed namespace config and init a newed namespace.
	// NOTICE: namespace must registered before start.
	Register(name string) error

	// DeRegister de-registers namespace from system by the given name.
	// NOTICE: DeRegister should be called after registering that namespace,
	// and we recommend stopping namespace first if you want to de-register a namespace.
	DeRegister(name string) error

	// GetNamespaceByName returns the namespace instance by name.
	GetNamespaceByName(name string) Namespace

	// StartNamespace starts namespace by name. This should only be called by
	// hypercli admin interface.
	StartNamespace(name string) error

	// StopNamespace stops namespace by name. This should only be called by
	// hypercli admin interface.
	StopNamespace(name string) error

	// RestartNamespace restarts namespace by name. This should only be called
	// by hypercli admin interface.
	RestartNamespace(name string) error

	// StartJvm starts jvm manager. This should only be called by hypercli
	// admin interface.
	StartJvm() error

	// StopJvm stops jvm manager. This should only be called by hypercli
	// admin interface.
	StopJvm() error

	// RestartJvm restarts jvm manager. This should only be called by
	// hypercli admin interface.
	RestartJvm() error

	// GlobalConfig returns the global configuration of the system.
	GlobalConfig() *common.Config

	// GetStopFlag returns the flag of stop hyperchain server
	GetStopFlag() chan bool

	// GetRestartFlag returns the flag of restart hyperchain server
	GetRestartFlag() chan bool

	// ProcessRequest dispatches received requests to corresponding namespace
	// processor.
	// Requests are sent from RPC layer, so responses are returned to RPC layer
	// with the certain namespace.
	ProcessRequest(namespace string, request interface{}) interface{}

	// GetNamespaceProcessor returns the namespace instance by name.
	GetNamespaceProcessor(name string) intfc.NamespaceProcessor

	InternalServer() server.InternalServer
}



// nsManagerImpl implements the NamespaceManager interface.
type nsManagerImpl struct {
	// Since param "namespaces" may be read/write in different go-routines,
	// it should be protected by a rwLock.
	rwLock *sync.RWMutex

	// namespaces stores all the namespace instances registered in the system
	namespaces map[string]Namespace

	// jvmManager manages the jvm executor in system, since executor is a
	// pluggable components, so start jvm manager or not should be written
	// in the config file.
	jvmManager *JvmManager

	// bloomfilter is the transaction bloom filter, helps to do transaction
	// duplication checking
	bloomFilter bloom.TxBloomFilter

	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	status *Status

	is *server.InternalServer

	stopHp    chan bool
	restartHp chan bool
}

// newNsManager news a namespace manager implement and init.
func newNsManager(conf *common.Config, stopHp chan bool, restartHp chan bool) *nsManagerImpl {
	status := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}

	//TODO: refactor this later
	var s *server.InternalServer
	if !conf.GetBool(common.EXECUTOR_EMBEDDED) {
		srv, err := server.NewInternalServer(conf.GetInt(common.INTERNAL_PORT), "0.0.0.0")
		if err != nil {
			panic(err)
		}
		lis, err := net.Listen("tcp", srv.Addr())
		if err != nil {
			panic(lis)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterDispatcherServer(grpcServer, srv)

		logger.Criticalf("InternalServer start successful on %v !", srv.Addr())
		go grpcServer.Serve(lis)
		s = srv
	}

	nr := &nsManagerImpl{
		namespaces:  make(map[string]Namespace),
		conf:        conf,
		jvmManager:  NewJvmManager(conf),
		bloomFilter: bloom.NewBloomFilterCache(conf),
		status:      status,
		stopHp:      stopHp,
		restartHp:   restartHp,
		is:          s,
	}

	nr.rwLock = new(sync.RWMutex)
	nr.bloomFilter.Start()
	return nr
}

// GetNamespaceManager returns the namespace manager instance.
// This function can only be called once, if called more than once,
// only returns the unique NamespaceManager.
func GetNamespaceManager(conf *common.Config, stopHp chan bool, restartHp chan bool) NamespaceManager {
	logger = common.GetLogger(common.DEFAULT_LOG, "nsmgr")
	once.Do(func() {
		nr = newNsManager(conf, stopHp, restartHp)
	})
	return nr
}


// init initializes the nsManagerImpl and retrieves the namespaces
// with the name of dirs under the NS_CONFIG_DIR_ROOT path, if the
// namespace name has been set true in config file, register the
// namespace, else, retrieves the next.
func (nr *nsManagerImpl) init() error {
	configRootDir := nr.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid ")
	}
	dirs, err := ioutil.ReadDir(configRootDir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if d.IsDir() {
			name := d.Name()
			start := nr.conf.GetBool(common.START_NAMESPACE + name)
			if !start {
				continue
			}
			// only register namespace whose name has been set true
			// in config file.
			if err := nr.Register(name); err != nil {
				logger.Error(err)
				return err
			}
		} else {
			logger.Errorf("Invalid folder %v", d)
		}
	}
	nr.status.setState(initialized)
	return nil
}

// Start starts namespace manager service which will start all namespaces
// registered in system and then start jvm manager if needed.
func (nr *nsManagerImpl) Start() error {
	state := nr.status.getState()
	if state < initialized {
		err := nr.init()
		if err != nil {
			logger.Errorf("Init namespace manager failed: %v", err)
			return err
		}
	}
	if nr.status.getState() == running {
		logger.Critical("namespace manager is already running")
		return nil
	}

	if !nr.conf.GetBool(common.EXECUTOR_EMBEDDED) {
		logger.Criticalf("waitting for executor admin to connect ...")
		//TODO: add timeout detect
		admin := <-nr.is.AdminRegister()
		logger.Criticalf("executor admin at %v connected", admin)
	}
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		go func(name string) {
			err := nr.StartNamespace(name)
			if err != nil {
				logger.Errorf("namespace %s start failed, %v", name, err)
			}
		}(name)
	}
	if nr.conf.GetBool(common.C_JVM_START) == true && nr.conf.GetBool(common.EXECUTOR_EMBEDDED) {
		if err := nr.jvmManager.Start(); err != nil {
			logger.Error(err)
			return err
		}
	}
	nr.status.setState(running)
	return nil
}

// Stop stops namespace manager service which will stop all namespaces
// registered in system and then stop jvm manager.
func (nr *nsManagerImpl) Stop() error {
	state := nr.status.getState()
	if state != running {
		logger.Critical("namespace manager is not running now, need not to stop")
		return nil
	}

	logger.Noticef("Try to stop NamespaceManager ...")
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		err := nr.StopNamespace(name)
		if err != nil {
			logger.Errorf("namespace %s stop failed, %v", name, err)
			return err
		}
	}
	if err := nr.jvmManager.Stop(); err != nil {
		logger.Errorf("Stop hyperjvm error %v", err)
	}
	nr.bloomFilter.Close()
	nr.status.setState(closed)
	logger.Noticef("NamespaceManager stopped!")
	return nil
}

// List returns all namespace names in system.
func (nr *nsManagerImpl) List() (names []string) {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		names = append(names, name)
	}
	return names
}

// checkNamespaceName checks the format of namespace name, it must be:
// global or other unique namespace name generated by hypercli whose length
// must be 13 and prefixed by "ns_".
func (nr *nsManagerImpl) checkNamespaceName(name string) bool {
	if name == "global" {
		return true
	}
	if len(name) == 13 && strings.HasPrefix(name, "ns_") {
		return true
	}
	return false
}



// Register registers a newed namespace to system by the newed namespace
// config dir and update the config file if needed.
func (nr *nsManagerImpl) Register(name string) error {
	logger.Noticef("Register namespace: %s", name)
	if _, ok := nr.namespaces[name]; ok {
		logger.Warningf("namespace [%s] has been registered", name)
		return ErrRegistered
	}
	configRootDir := nr.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid ")
	}
	nsRootPath := configRootDir + "/" + name
	if _, err := os.Stat(nsRootPath); os.IsNotExist(err) {
		logger.Errorf("namespace [%s] root path doesn't exist!", name)
		return ErrNonExistConfig
	}
	nsConfigDir := nsRootPath + "/config"
	nsConfig, err := nr.constructConfigFromDir(name, nsConfigDir)
	if err != nil {
		return err
	}
	nsConfig.Set(common.JSON_RPC_PORT, "8081")
	nsConfig.Set(common.EXECUTOR_HOST_ADDR, "127.0.0.1")
	delFlag := make(chan bool)
	ns, err := GetNamespace(name, nsConfig, delFlag, nr.is)
	if nr.conf.GetBool(common.EXECUTOR_EMBEDDED) == false {
		nr.is.RegisterLocal(ns.LocalService()) // register local service
	}
	if err != nil {
		logger.Errorf("Construct namespace %s error, %v", name, err)
		return ErrCannotNewNs
	}
	nr.addNamespace(ns)
	if nr.conf.GetBool(common.EXECUTOR_EMBEDDED) {
		if err := nr.bloomFilter.Register(name); err != nil {
			logger.Error("register bloom filter failed", err.Error())
			return err
		}
	}

	if err = updateNamespaceStartConfig(name, nr.conf); err != nil {
		logger.Criticalf("Update namespace start for [%s] config failed", name)
	}
	go nr.ListenDelNode(name, delFlag)
	return nil
}

// DeRegister de-registers namespace from system by name.
func (nr *nsManagerImpl) DeRegister(name string) error {
	logger.Criticalf("Try to deregister the namespace:%s ", name)
	if ns, ok := nr.namespaces[name]; ok {
		if ns.Status().state == running {
			logger.Noticef("namespace: %s is running, stop it first", name)
			ns.Stop()
		}
		nr.removeNamespace(name)
	} else {
		logger.Warningf("namespace %s not exist, please register first.", name)
		return ErrNoSuchNamespace
	}
	nr.bloomFilter.UnRegister(name)
	logger.Criticalf("namespace: %s stopped", name)
	//TODO: need to delete the data and stop listen del node.
	return nil
}

// updateNamespaceStartConfig updates the config file with the specified name
func updateNamespaceStartConfig(name string, conf *common.Config) error {
	if !conf.ContainsKey(fmt.Sprintf("namespace.start.%s", name)) {
		return common.SeekAndAppend("[namespace.start]", conf.GetString(common.GLOBAL_CONFIG_PATH),
			fmt.Sprintf("    %s = true", name))
	}
	return nil
}

// removeNamespace removes the namespace instance with the given
// name from system.
func (nr *nsManagerImpl) removeNamespace(name string) {
	nr.rwLock.Lock()
	delete(nr.namespaces, name)
	nr.rwLock.Unlock()
}

// addNamespace adds the namespace instance with the given name to system.
func (nr *nsManagerImpl) addNamespace(ns Namespace) {
	nr.rwLock.Lock()
	nr.namespaces[ns.Name()] = ns
	nr.rwLock.Unlock()
}

// GetNamespaceByName get namespace instance by name.
func (nr *nsManagerImpl) GetNamespaceByName(name string) Namespace {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		return ns
	}
	return nil
}

func (nr *nsManagerImpl)  GetNamespaceProcessor(name string) intfc.NamespaceProcessor {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		return ns
	}
	return nil
}

// ProcessRequest dispatches the request to the specified namespace processor.
func (nr *nsManagerImpl) ProcessRequest(namespace string, request interface{}) interface{} {
	np := nr.GetNamespaceProcessor(namespace)
	if np == nil {
		logger.Noticef("no namespace found for name: %s", namespace)
		return nil
	}
	return np.ProcessRequest(request)
}

// StartNamespace starts namespace instance by name.
func (nr *nsManagerImpl) StartNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		if err := ns.Start(); err != nil {
			ns.Stop() //start failed, try to stop some started components
			return err
		} else {
			if nr.conf.GetBool(common.EXECUTOR_EMBEDDED) {
				nr.jvmManager.ledgerProxy.RegisterDB(name, ns.GetExecutor().FetchStateDb())
			}
			return nil
		}
	}
	logger.Errorf("No namespace instance for %s found", name)
	return ErrInvalidNs
}

// StopNamespace stops namespace instance by name.
func (nr *nsManagerImpl) StopNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		if err := ns.Stop(); err != nil {
			return err
		} else {
			nr.jvmManager.ledgerProxy.UnRegister(name)
			return nil
		}
	}
	logger.Errorf("No namespace instance for %s found")
	return ErrInvalidNs
}

// RestartNamespace restarts namespace by name.
func (nr *nsManagerImpl) RestartNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		return ns.Restart()
	}
	logger.Errorf("No namespace instance for %s found")
	return ErrInvalidNs
}

// GlobalConfig returns the global configuration of the system.
func (nr *nsManagerImpl) GlobalConfig() *common.Config {
	return nr.conf
}

// StartJvm starts the jvm manager.
func (nr *nsManagerImpl) StartJvm() error {
	if err := nr.jvmManager.Start(); err != nil {
		return err
	}
	return nil
}

// StopJvm stops the jvm manager.
func (nr *nsManagerImpl) StopJvm() error {
	if err := nr.jvmManager.Stop(); err != nil {
		return err
	}
	return nil
}

// RestartJvm restarts the jvm manager
func (nr *nsManagerImpl) RestartJvm() error {
	if err := nr.jvmManager.Stop(); err != nil {
		return err
	}
	if err := nr.jvmManager.Start(); err != nil {
		return err
	}
	return nil
}

// ListenDelNode listens delete node event from p2p module.
func (nr *nsManagerImpl) ListenDelNode(name string, delFlag chan bool) {
	for {
		select {
		case <-delFlag:
			nr.StopNamespace(name)
			nr.DeRegister(name)
		}
	}
}

// GetStopFlag returns the flag of stop hyperchain server
func (nr *nsManagerImpl) GetStopFlag() chan bool {
	return nr.stopHp
}

// GetRestartFlag returns the flag of restart hyperchain server
func (nr *nsManagerImpl) GetRestartFlag() chan bool {
	return nr.restartHp
}

func (nr *nsManagerImpl) InternalServer() *server.InternalServer {
	return nr.is
}
