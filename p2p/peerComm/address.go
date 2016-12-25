//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

type Address struct {
	ID      int
	Port    int
	RPCPort int
	IP      string
}

func NewAddress(id int, port int, rpcPort int, ip string) Address {
	address := Address{
		ID:      id,
		Port:    port,
		RPCPort: rpcPort,
		IP:      ip,
	}
	return address
}
