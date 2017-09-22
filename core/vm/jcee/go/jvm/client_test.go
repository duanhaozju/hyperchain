package jvm

import (
	"hyperchain/common"
	pb "hyperchain/core/vm/jcee/protos"
	"testing"
)

func NewJvmClient() *Client {
	conf := common.NewRawConfig()
	conf.Set(common.JVM_PORT, 50051)

	client := &Client{
		config: conf,
	}
	return client
}

func TestClient_SyncExecute(t *testing.T) {
	client := NewJvmClient()
	err := client.Connect()

	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 100; i++ {
		rsp, err := client.SyncExecute(&pb.Request{
			Method: "test",
		})
		if err != nil {
			t.Error(err)
		}

		t.Log(rsp)
	}
}
