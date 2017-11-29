package fiber

import "github.com/hyperchain/hyperchain/manager/protos"

//Fiber response for transferring data between distributed components.
type Fiber interface {
	//Start start the fiber system
	Start() error
	//Stop the fiber system
	Stop()
	//FetchBlockchainInfo
	FetchBlockchainInfo() chan *protos.BlockchainInfo // block chain info contains checkpoint info
	//Send info to the remote peer
	Send(msg interface{}) error
}
