//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"errors"
	"fmt"

	"hyperchain/core/vm/evm/params"
)

var OutOfGasError = errors.New("Out of gas")
var CodeStoreOutOfGasError = errors.New("Contract creation code storage out of gas")
var DepthError = fmt.Errorf("Max call depth exceeded (%d)", params.CallCreateDepth)