//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"encoding/base64"
)

// LoginServer implements the basic auth login.
func (admin *Administrator) LoginServer(w http.ResponseWriter, req *http.Request) {
	admin.logger.Debug("start listen login function...")

	// get username and password from http request.
	username, password, err := admin.basicAuth(req)
	if err != nil {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized!")
		return
	}

	// judge if the user exist or not.
	if ok, err := admin.isUserExist(username, password); !ok {
		admin.logger.Debugf("User %s: %s", username, err.Error())
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, err.Error())
		return
	}

	// generate signed token.
	token, err := signToken(username, pri_key, "RS256")
	if err != nil {
		admin.logger.Error("Failed to generate signed token:", err)
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Error in generate token:%s", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, err.Error())
		return
	}

	// update last operation time of this user.
	admin.updateLastOperationTime(username)
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("%s", LoggedIn))
}

// basicAuth decodes username and password from http request and returns
// an error if encountered an error.
func (admin *Administrator) basicAuth(req *http.Request) (string, string, error) {
	var username, password string
	auth := req.Header.Get("Authorization")
	if auth == "" {
		admin.logger.Debug("Need authorization field in http head.")
		return "", "", ErrNotLogin
	}
	auths := strings.SplitN(auth, " ", 2)
	if len(auths) != 2 {
		admin.logger.Debug("Login params need to be split with blank.")
		return "", "", ErrNotLogin
	}
	authMethod := auths[0]
	authB64 := auths[1]
	switch authMethod {
	case "Basic":
		authstr, err := base64.StdEncoding.DecodeString(authB64)
		if err != nil {
			admin.logger.Debug(err)
			return "", "", ErrDecodeErr
		}
		userPwd := strings.SplitN(string(authstr), ":", 2)
		if len(userPwd) != 2 {
			admin.logger.Debug("username and password must be split by ':'")
			return "", "", ErrDecodeErr
		}
		username = userPwd[0]
		password = userPwd[1]
		return username, password, nil

	default:
		admin.logger.Notice("Not supported auth method")
		return "", "", ErrNotSupport
	}
}
