//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"hyperchain/api/admin"
)

// CmdClient manages a connection to the host:port which may be used once in
// one cli command.
type CmdClient struct {
	// host name (ip)
	host string

	// host port
	port string

	// http client instance
	client *http.Client
}

// NewRpcClient returns a new client connection to the given host and port.
func NewRpcClient(host, port string) *CmdClient {
	cc := &CmdClient{
		host: host,
		port: port,
	}
	client := &http.Client{}
	cc.client = client
	return cc
}

//InvokeCmd invokes a command using json-format request and waits for the response.
func (cc *CmdClient) InvokeCmd(cmd *admin.Command) string {
	rs, err := cc.Call(cmd.ToJson(), cmd.MethodName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return ""
	}
	return rs
}

// Call actually sends Command to hyperchain server using the given cmd.
func (cc *CmdClient) Call(cmd string, method string) (string, error) {
	reqJson := []byte(cmd)
	urlStr := fmt.Sprintf("http://%s:%s", cc.host, cc.port)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqJson))
	if err != nil {
		return "", err
	}

	// get authorization token
	userinfo := new(UserInfo)
	ReadFile(tokenpath, userinfo)
	req.Header.Set("Authorization", userinfo.Token)
	req.Header.Set("Method", method)

	rs, err := cc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return "", err
	}
	result := string(body)

	// check if returns a token error
	if err = checkToken(result); err != nil {
		return "", err
	}
	return result, nil
}

// Login actually sends login request to hyperchain server with the given username
// and password.
func (cc *CmdClient) Login(username, password string) (string, error) {
	reqJson := []byte("")
	urlStr := fmt.Sprintf("http://%s:%s%s", cc.host, cc.port, "/login")
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqJson))
	if err != nil {
		return "", err
	}
	// set basic auth header
	req.SetBasicAuth(username, password)

	rs, err := cc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return "", err
	}
	result := string(body)
	if rs.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf(result)
	}
	fmt.Println(result)

	// token will be stored in Authorization head
	token := rs.Header.Get("Authorization")
	if token == "" {
		return "", ErrEmptyHeader
	} else {
		return token, nil
	}
}
