package ipc

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

var (
	logger               *logging.Logger
	tcpKeepAliveInterval = 30 * time.Second
)

type IPCServer struct {
	endpoint string
}

func NEWIPCServer(endpoint string) *IPCServer {
	logger = common.GetLogger(common.DEFAULT_LOG, "ipc")
	return &IPCServer{
		endpoint: endpoint,
	}
}

func (server *IPCServer) Start() error {
	var (
		listener net.Listener
		err      error
	)

	rpc.Register(GetRemoteCall())
	rpc.HandleHTTP()

	logger.Notice("interactive ipc shell server listening...")
	if listener, err = server.listener(); err != nil {
		logger.Errorf("some error occured: %s", err.Error())
		return err
	}

	go func() {
		err = http.Serve(listener, nil)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (server *IPCServer) listener() (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(server.endpoint), 0751); err != nil {
		return nil, err
	}

	os.Remove(server.endpoint)
	logger.Noticef("start unix ipc server, pipe: %s", server.endpoint)
	l, err := net.Listen("unix", server.endpoint)
	if err != nil {
		return nil, err
	}
	os.Chmod(server.endpoint, 0600)
	return l, nil
}
