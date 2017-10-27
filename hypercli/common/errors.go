//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import "errors"

var (
	ErrInvalidParams  = errors.New("hypercli/cmd: invalid params")
	ErrInvalidArgsNum = errors.New("hypercli/cmd: invalid args num")
	ErrEmptyHeader    = errors.New("empty authorization header.")
	ErrDecrypt        = errors.New("could not decrypt key with given passphrase")
)
