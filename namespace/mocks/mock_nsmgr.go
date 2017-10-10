package mocks

import (
	"github.com/stretchr/testify/mock"
	"hyperchain/common"
	"hyperchain/namespace"
)

type MockNSMgr struct {
	mock.Mock
}

func (mnsm *MockNSMgr) Start() error {
	args := mnsm.Called()
	return args.Error(0)
}

func (mnsm *MockNSMgr) Stop() error {
	args := mnsm.Called()
	return args.Error(0)
}

func (mnsm *MockNSMgr) List() []string {
	args := mnsm.Called()
	return args.Get(0).([]string)
}

func (mnsm *MockNSMgr) Register(name string) error {
	args := mnsm.Called(name)
	return args.Error(0)
}

func (mnsm *MockNSMgr) DeRegister(name string) error {
	args := mnsm.Called(name)
	return args.Error(0)
}

func (mnsm *MockNSMgr) GetNamespaceByName(name string) namespace.Namespace {
	args := mnsm.Called(name)
	return args.Get(0).(namespace.Namespace)
}

func (mnsm *MockNSMgr) ProcessRequest(namespace string, request interface{}) interface{} {
	args := mnsm.Called(namespace, request)
	return args.Get(0)
}

func (mnsm *MockNSMgr) StartNamespace(name string) error {
	args := mnsm.Called(name)
	return args.Error(0)
}

func (mnsm *MockNSMgr) StopNamespace(name string) error {
	args := mnsm.Called(name)
	return args.Error(0)
}

func (mnsm *MockNSMgr) RestartNamespace(name string) error {
	args := mnsm.Called(name)
	return args.Error(0)
}

func (mnsm *MockNSMgr) StartJvm() error {
	args := mnsm.Called()
	return args.Error(0)
}

func (mnsm *MockNSMgr) StopJvm() error {
	args := mnsm.Called()
	return args.Error(0)
}

func (mnsm *MockNSMgr) RestartJvm() error {
	args := mnsm.Called()
	return args.Error(0)
}

func (mnsm *MockNSMgr) GlobalConfig() *common.Config {
	args := mnsm.Called()
	return args.Get(0).(*common.Config)
}

func (mnsm *MockNSMgr) GetStopFlag() chan bool {
	args := mnsm.Called()
	return args.Get(0).(chan bool)
}

func (mnsm *MockNSMgr) GetRestartFlag() chan bool {
	args := mnsm.Called()
	return args.Get(0).(chan bool)
}
