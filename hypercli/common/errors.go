//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import "errors"

var (
	ErrInvalidParams  = errors.New("hypercli/cmd: invalid params")
	ErrInvalidArgsNum = errors.New("hypercli/cmd: invalid args num")
	ErrInvalidCmd     = errors.New("hypercli/cmd: invalid command")
	ErrEmptyHeader    = errors.New("Empty authorization header.")
	ErrInvalidToken   = errors.New("Invalid token, please login first!")
	ErrDecrypt        = errors.New("could not decrypt key with given passphrase")
)
