package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/hyperchain/hyperchain/p2p/message"
	"google.golang.org/grpc/metadata"
	"golang.org/x/net/context"
)

type MockChat_ChatServer struct {
	mock.Mock
}

func (server *MockChat_ChatServer) Send(msg *message.Message) error {
	args := server.Called(msg)
	return args.Error(0)
}

func (server *MockChat_ChatServer) Recv() (*message.Message, error) {
	args := server.Called()
	return args.Get(0).(*message.Message), args.Error(1)
}

func (server *MockChat_ChatServer) SetHeader(md metadata.MD) error {
	args := server.Called(md)
	return args.Error(0)
}

func (server *MockChat_ChatServer) SendHeader(md metadata.MD) error {
	args := server.Called(md)
	return args.Error(0)
}

func (server *MockChat_ChatServer) SetTrailer(metadata.MD) {}

func (server *MockChat_ChatServer) Context() context.Context {
	args := server.Called()
	return args.Get(0).(context.Context)
}

func (server *MockChat_ChatServer) SendMsg(m interface{}) error {
	args := server.Called(m)
	return args.Error(0)
}

func (server *MockChat_ChatServer) RecvMsg(m interface{}) error {
	args := server.Called(m)
	return args.Error(0)
}

