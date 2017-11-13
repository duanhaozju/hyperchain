//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	admin "hyperchain/api/admin"
	"io/ioutil"
	"net/http"
	"os"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("hypercli/common")
}

func GetCmdClient(c *cli.Context) *CmdClient {
	client := NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	return client
}

func GetCmdRemoteClient(c *cli.Context) *CmdClient {
	client := NewRpcClient(c.GlobalString("exehost"), c.GlobalString("exeport"))
	return client
}

type CmdClient struct {
	host   string
	port   string
	client *http.Client
}

func (cc *CmdClient) init() {
	client := &http.Client{}
	cc.client = client
}

func NewRpcClient(host, port string) *CmdClient {
	client := &CmdClient{
		host: host,
		port: port,
	}
	client.init()
	return client
}

//InvokeCmd invoke a command using json admin client and wait for the response.
func (cc *CmdClient) InvokeCmd(cmd *admin.Command) string {
	rs, err := cc.Call(cmd.ToJson(), cmd.MethodName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return ""
	}
	return rs
}

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
	if err = checkToken(result); err != nil {
		return "", err
	}
	return result, nil
}

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
	token := rs.Header.Get("Authorization")
	if token == "" {
		return "", ErrEmptyHeader
	} else {
		return token, nil
	}
}
