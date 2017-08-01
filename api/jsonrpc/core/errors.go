package jsonrpc

import (
	"errors"
)

// Error constants
var (
	ErrNotLogin          = errors.New("Need User Login")
	ErrUnMatch           = errors.New("Username or password invalid")
	ErrUserNotExist      = errors.New("User doesn't exist")
	ErrDecodeErr         = errors.New("Decode error")
	ErrNotSupport        = errors.New("Not support method")
	ErrNoAuth            = errors.New("Failed to verify your signature")
	ErrPermission        = errors.New("Permission denied")
	ErrTimeoutPermission = errors.New("Expired token, please login again")
	ErrInternal          = errors.New("Internal error")
	ErrTokenInvalid      = errors.New("Invalid token, please login first")
	ErrDuplicateUsername = errors.New("Duplicate username during register, please register with a new username")
	ErrInvalidGroup      = errors.New("Unrecognized group, please contact to the administrator")
)
