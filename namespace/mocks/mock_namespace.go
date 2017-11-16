package mocks

import (
	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/core/executor"
	"github.com/hyperchain/hyperchain/namespace"
	"github.com/stretchr/testify/mock"
)

type MockNS struct {
	mock.Mock
}

func (m *MockNS) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNS) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNS) Restart() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNS) Status() *namespace.Status {
	args := m.Called()
	return args.Get(0).(*namespace.Status)
}

func (m *MockNS) ProcessRequest(request interface{}) interface{} {
	args := m.Called(request)
	return args.Get(0)
}

func (m *MockNS) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockNS) GetCAManager() *admittance.CAManager {
	args := m.Called()
	return args.Get(0).(*admittance.CAManager)
}

func (m *MockNS) GetExecutor() *executor.Executor {
	args := m.Called()
	return args.Get(0).(*executor.Executor)
}
