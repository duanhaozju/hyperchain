package message

// this file is assent file for endpoint.pb.go
//NewEndpoint return a new endpoint infomation
func NewEndpoint(hostname string, namespace string, hash string) *Endpoint {
	return &Endpoint{
		Version:  MSG_ENDPOINT_VERSION,
		Hostname: []byte(hostname),
		Field:    []byte(namespace),
		UUID:     []byte(hash),
	}
}
