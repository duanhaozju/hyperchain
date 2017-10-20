package apiserver

type APIServer interface {
	Start()
	Stop()

}

type APIServerImpl struct{
	//rpcServer *jsonrpc.RPCServerImpl
}

//func NewAPIServer(nr namespace.NamespaceManager, config *common.Config) *APIServerImpl{
//	//s := jsonrpc.GetRPCServer(nr, config)
//	//apiserver := &APIServerImpl{
//	//	rpcServer: s,
//	//}
//	//return apiserver
//}

func(asi *APIServerImpl) Start() error{
	//return asi.rpcServer.Start()
	return nil
}

func(asi *APIServerImpl) Stop() error{
	return asi.Stop()
}
// TODO : implement it
