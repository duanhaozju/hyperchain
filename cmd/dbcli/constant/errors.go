package constant

import "errors"

var (
	ErrInvalidParams = errors.New("dbcli/cmd: invalid params.")
	ErrDBInit        = errors.New("dbcli/cmd: database initilize fail.")
	ErrQuery         = errors.New("dbcli/cmd: query database fail.")
	ErrCreateDBCopy  = errors.New("dbcli/cmd: create database copy fail.")

	ErrInvalidDBParams = errors.New("dbcli/cmd: invalid db params.")
	ErrDataVersion     = errors.New("data version error!")
	ErrDataType        = errors.New("data struct type error!")
)
