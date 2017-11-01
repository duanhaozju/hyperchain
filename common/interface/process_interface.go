package intfc

import "hyperchain/admittance"

type NsMgrProcessor interface {

	// ProcessRequest dispatches received requests to corresponding namespace
	// processor.
	// Requests are sent from RPC layer, so responses are returned to RPC layer
	// with the certain namespace.
	ProcessRequest(namespace string, request interface{}) interface{}

	// GetNamespaceProcessor returns the namespace processor instance by name.
	GetNamespaceProcessor(name string) NamespaceProcessor
}

type NamespaceProcessor interface {

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	GetCAManager() *admittance.CAManager
}
