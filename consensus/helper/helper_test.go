// author: Xiaoyi Wang
// email: wangxiaoyi@hyperchain.cn
// date: 16/11/1
// last modified: 16/11/1
// last Modified Author: Xiaoyi Wang
// change log: 1. new test for helper

package helper

import (
	"testing"

	"hyperchain/event"
	"reflect"
)

func TestNewHelper(t *testing.T)  {
	m := &event.TypeMux{}
	h := &helper{
		msgQ: m,
	}

	help := NewHelper(m)
	if !reflect.DeepEqual(help, h) {
		t.Errorf("error NewHelper")
	}
}