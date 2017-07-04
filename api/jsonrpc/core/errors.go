package jsonrpc

import (
	"errors"
)

// Error constants
var (
	ErrNotLogin          = errors.New("Need User Login")
	ErrUnMatch           = errors.New("Username does't match the password")
	ErrUserNotExist      = errors.New("User doesn't exist")
	ErrDecodeErr         = errors.New("Decode error")
	ErrNotSupport        = errors.New("Not support method")
	ErrNoAuth            = errors.New("Failed to verify your signature")
	ErrPermission        = errors.New("Permission denied")
	ErrInternal          = errors.New("Internal error")
	ErrTokenInvalid      = errors.New("Invalid token, please login first")
	ErrDuplicateUsername = errors.New("Duplicate username during register, please register with a new username")
)
