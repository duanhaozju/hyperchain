//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package admin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignToken(t *testing.T) {
	var username, algorithm, keypath string
	var err error
	ast := assert.New(t)

	// mock: sign token with an invalid key path
	wrongPath := "./key"
	_, err = signToken(username, wrongPath, algorithm)
	ast.Contains(err.Error(), "Couldn't read key",
		"Failed to sign token, as we provide an invalid path.")

	// mock: sign token with an invalid algorithm
	wrongAlg := "RS257"
	_, err = signToken(username, keyPath, wrongAlg)
	ast.Contains(err.Error(), "Couldn't find signing method",
		"Failed to sign token, as we provide an invalid algorithm.")

	// mock: sign token with correct args
	username = "hpc"
	algorithm = "RS256"
	keypath = keyPath
	_, err = signToken(username, keypath, algorithm)
	ast.Equal(nil, err, "Successfully sign token, should return nil error")
}

func TestVerifyToken(t *testing.T) {
	var auth, keypath, algorithm string
	var err error
	ast := assert.New(t)

	// mock: verify token with an invalid key path
	wrongPath := "./key"
	_, err = verifyToken(auth, wrongPath, algorithm)
	ast.Contains(err.Error(), "Invalid token",
		"Failed to verify token, as we provide an invalid path.")

	// mock: verify token with an invalid algorithm
	wrongAlg := "RS257"
	_, err = verifyToken(auth, keypath, wrongAlg)
	ast.Contains(err.Error(), "Invalid token",
		"Failed to verify token, as we provide an invalid algorithm.")

	// mock: sign token with correct args.
	username := "hpc"
	token, err := signToken(username, keyPath, "RS256")
	t.Logf("token is : %v", token)
	keypath = "../../hypercli/keyconfigs/key/key.pub"
	algorithm = "RS256"
	claim, err := verifyToken(token, keypath, algorithm)
	ast.Equal(nil, err,
		"Successfully verify token, as we provide an correct args.")

	user := getUserFromClaim(claim)
	ast.Equal(username, user, "They should bre equal.")
}
