//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.setNormal()
	rbft.setFull()
	n, f := rbft.GetStatus()
	ast.Equal(true, n, "GetStatus failed")
	ast.Equal(true, f, "GetStatus failed")

	rbft.setNotFull()
	_, f = rbft.GetStatus()
	ast.Equal(false, f, "GetStatus failed")
}
