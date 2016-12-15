//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	"testing"
	"math/big"
)

var(
	num1 = big.NewInt(2)
	num2 = big.NewInt(3)
	num3 = big.NewInt(4)
)

func TestStack(t *testing.T) {
	stack := newstack()
	stack.push(num1)
	d := stack.len()
	if d!=1{
		t.Error("wrong stack push")
	}
	stack.pushN(num2,num3)
	if d=stack.len();d!=3{
		t.Error("wrong stack pushN")
	}
	ret := stack.pop()
	if ret!=num3{
		t.Error("wrong stack pop")
	}
	data := stack.Data()
	if(len(data)!=2){
		t.Error("wrong stack Data method")
	}
	peek := stack.peek()
	if peek!=num2 {
		t.Error("wrong stack peek")
	}
	if err:=stack.require(3);err==nil{
		t.Error("stack require error")
	}
	if err:= stack.require(1);err!=nil{
		t.Error("stack requier error")
	}
	stack.pushN(num1,num2,num3)
	//now the data in stack:[2,3,2,3,4]
	stack.dup(1)
	if (stack.len()!=6){
		t.Error("stack dup error")
		tmp := stack.peek()
		if tmp.Cmp(num3)!=0{
			t.Error("stack dup error")
		}
	}
	stack.swap(3)
	stack.Print()
	//now stack is [2 3 2 4 4 3]
	len := stack.len()
	swp1 := stack.data[len-1]
	swp2 := stack.data[len-3]
	if swp1.Cmp(num2)!=0 || swp2.Cmp(num3)!=0{
		t.Error("stack swap error")
	}
	empty := newstack()
	empty.Print()

}
