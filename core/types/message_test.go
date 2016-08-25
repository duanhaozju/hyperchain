// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"testing"

	"fmt"

	"math/big"
)

type Transaction_test struct {
	//From      []byte
	From      *big.Int
	//From     string


}

// This test should panic if concurrent close isn't implemented correctly.
func TestMsgPipeConcurrentClose(t *testing.T) {
	/*tx:=&Transaction{
		From: "1",
		To: "2",
		Value: 1,
		TimeStamp: time.Now().Unix(),
	}*/

	tx :=new(Transaction_test)
	from:=new(big.Int)
	from,k:=from.SetString("123",10)

	//1.(*big.Int)
	fmt.Println(k)

	//tx.From=[]byte{0x00, 0x00, 0x03, 0xe8}
	tx.From=from


	//fmt.Println(tx.Hash())
	var txs []*Transaction_test
	txs = append(txs, tx)

	msg:=encode(1,tx)
	fmt.Println("msg")
	fmt.Println(msg)
	var txs1 *Transaction_test
	msg.Decode(&txs1)
	fmt.Println(tx)
	fmt.Println(txs1)

}



