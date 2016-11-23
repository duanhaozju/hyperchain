//author :zsx
//data: 2016-11-3

//publicNodeAPI.GetNodes()要获取节点信息，要启动的模块太多，放在集成测试中测试
package hpc

import (
	"testing"
)

func Test_GetNodes(t *testing.T) {
	publicNodeAPI := &PublicNodeAPI{}
	_, err := publicNodeAPI.GetNodes()
	if err == nil {
		t.Errorf("publicNodeAPI.GetNodes()")
	}
}
