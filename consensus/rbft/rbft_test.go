//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	ast := assert.New(t)
	rbft := new(rbftImpl)

	rbft.normal = uint32(1)
	rbft.poolFull = uint32(1)
	n, f := rbft.GetStatus()
	ast.Equal(true, n, "GetStatus failed")
	ast.Equal(true, f, "GetStatus failed")

	rbft.poolFull = uint32(0)
	_, f = rbft.GetStatus()
	ast.Equal(false, f, "GetStatus failed")
}
