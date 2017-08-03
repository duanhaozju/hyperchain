//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package primitives

import (
	"crypto/elliptic"
)

var (
	defaultCurve elliptic.Curve
)

// GetDefaultCurve returns the default elliptic curve used by the crypto layer
func GetDefaultCurve() elliptic.Curve {
	return defaultCurve
}
