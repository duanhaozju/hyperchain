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

// LoginServer implements the basic auth login
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
	if ok, err := IsUserExist(username, password); !ok {
		log.Debugf("User %s: %s", username, err.Error())
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, err.Error())
		return
	}

	// generate token
	token, err := signToken(username, pri_key, "RS256")
	if err != nil {
		log.Error("Failed to generate signed token:", err)
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Error in generate token:%s", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, err.Error())
		return
	}
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("%s", LoggedIn))
}

func newJsonHttpHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxHTTPRequestContentLength {
			http.Error(w,
				fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
				http.StatusRequestEntityTooLarge)
			return
		}
		auth := r.Header.Get("Authorization")

		// verify signed token
		if claims, err := verifyToken(auth, pub_key, "RS256"); err != nil {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, fmt.Sprintf("%s", err.Error()))
			return
		} else {
			var method = r.Header.Get("Method")
			if method == "" {
				io.WriteString(w, "Invalid request")
				return
			}
			if ok, err := checkPermission(claims, method); !ok {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic releam=%s", err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, fmt.Sprintf("%s", err.Error()))
				return
			}
		}

		w.Header().Set("content-type", "application/json")
		codec := NewJSONCodec(&httpReadWrite{r.Body, w}, r, srv.namespaceMgr)
		defer codec.Close()
		srv.ServeSingleRequest(codec, OptionMethodInvocation)
	}
}

