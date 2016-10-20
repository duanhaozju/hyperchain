package peerComm

type Address struct {
	ID      uint64
	Port    int64
	RPCPort int64
	IP      string
}

func NewAddress(id int64, port int64, rpcPort int64, ip string) Address {
	address := Address{
		ID:      uint64(id),
		Port:    port,
		RPCPort: rpcPort,
		IP:      ip,
	}
	return address
}
