//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMisc(t *testing.T) {
	a := Big("10")
	b := Big("57896044618658097711785492504343953926634992332820282019728792003956564819968")
	c := []byte{1, 2, 3, 4}

	assert.Equal(t, true, BitTest(a, 1), "Expected true got false")

	U256(a)
	S256(a)

	U256(b)
	S256(b)

	BigD(c)
}

func TestBigMax(t *testing.T) {
	a := Big("10")
	b := Big("5")

	max1 := BigMax(a, b)
	assert.Equal(t, a, max1, fmt.Sprintf("Expected %d got %d", a, max1))

	max2 := BigMax(b, a)
	assert.Equal(t, a, max2, fmt.Sprintf("Expected %d got %d", a, max2))
}

func TestBigMin(t *testing.T) {
	a := Big("10")
	b := Big("5")

	min1 := BigMin(a, b)
	assert.Equal(t, b, min1, fmt.Sprintf("Expected %d got %d", b, min1))

	min2 := BigMin(b, a)
	assert.Equal(t, b, min2, fmt.Sprintf("Expected %d got %d", b, min2))
}

func TestBigCopy(t *testing.T) {
	a := Big("10")
	b := BigCopy(a)
	c := Big("1000000000000")
	y := BigToBytes(b, 16)
	ybytes := []byte{0, 10}
	z := BigToBytes(c, 16)
	zbytes := []byte{232, 212, 165, 16, 0}

	assert.Equal(t, 0, bytes.Compare(y, ybytes), fmt.Sprintf("Got", ybytes))
	assert.Equal(t, 0, bytes.Compare(z, zbytes), fmt.Sprintf("Got", zbytes))
}

func TestFirstBitSet(t *testing.T) {
	a := Big("8")
	res := FirstBitSet(a)
	assert.Equal(t, 3, res, fmt.Sprintf("Got", res))

	b := Big("0")
	res = FirstBitSet(b)
	assert.Equal(t, 0, res, fmt.Sprintf("Got", res))
}
