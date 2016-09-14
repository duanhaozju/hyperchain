package jsonrpc

import (
	"testing"
	"fmt"
)

func TestServer(t *testing.T) {
	err := Start(8084,nil)

	fmt.Println(err)
}