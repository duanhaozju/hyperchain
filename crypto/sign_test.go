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

package crypto

import (

	"fmt"
	"testing"

	"math/big"

	"hyperchain-alpha/common"
)

func TestBox(t *testing.T) {

	key, err := GenerateKey()
	SaveECDSA("./testFile",key)
	priv,err:=LoadECDSA("./testFile")
	pub := key.PublicKey


	fmt.Println("public key is :")
	fmt.Println(pub)
	fmt.Println("private key is :")
	fmt.Println(key)
	tx, _ := NewTransaction(uint64(2), common.Address{}, big.NewInt(100), big.NewInt(100), big.NewInt(int64(2)), nil).SignECDSA(priv)
	//
	//common.HexToAddress()


	from, err := tx.From()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(from.Hex())

	fmt.Println(from)






}
