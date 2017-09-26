//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
//"fmt"
//"io"
//"net/http"
//"io"
//"fmt"
)
import (
	"fmt"
	"io"
	"net/http"
)

// LoginServer implements the basic auth login.
func LoginServer(w http.ResponseWriter, req *http.Request) {
	log.Debug("start listen login function...")

	// get username and password from http request
	username, password, err := basicAuth(req)
	if err != nil {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized!")
		return
	}

	// judge if the user exist or not
	if ok, err := isUserExist(username, password); !ok {
		log.Debugf("User %s: %s", username, err.Error())
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, err.Error())
		return
	}

	// generate signed token
	token, err := signToken(username, pri_key, "RS256")
	if err != nil {
		log.Error("Failed to generate signed token:", err)
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Error in generate token:%s", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, err.Error())
		return
	}

	// update last operation time of this user.
	updateLastOperationTime(username)
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("%s", LoggedIn))
}
