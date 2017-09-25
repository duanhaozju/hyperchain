package dispatcher

import (
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
)

//DispatchServer dispatch service
type DispatchServer struct {
	port int
	host string
	sr   common.ServiceRegistry
}

func NewDispatchServer(port int, host string) (*DispatchServer, error) {
	ds := &DispatchServer{
		port: port,
		host: host,
	}
	return ds, nil
}

//Register receive a new connection
func (ds *DispatchServer) Register(stream pb.Dispatcher_RegisterServer) error {

	return nil
}

//dispatch dispatch messages
func (ds *DispatchServer) dispatch() {

}


