package network

import (
	"testing"
	"hyperchain/p2p/utils"
)

func TestGetInnerAddr(t *testing.T) {
	t.Log(utils.GetProjectPath()+"/p2p/test/addr.yaml")
	_,err := GetInnerAddr(utils.GetProjectPath()+"/p2p/test/addr.yaml")
	if err != nil{
		t.Fatal(err)
	}
}


func TestInnerAddr_Serialize(t *testing.T) {
	t.Log(utils.GetProjectPath()+"/p2p/test/addr.yaml")
	addr,err := GetInnerAddr(utils.GetProjectPath()+"/p2p/test/addr.yaml")
	if err != nil{
		t.Fatal(err)
	}
	b,e := addr.Serialize()
	if e != nil{
		t.Fatal(e)
	}
	t.Log(string(b))
}

