//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"net/http"
	"io/ioutil"
	"github.com/urfave/cli"
	admin "hyperchain/api/jsonrpc/core"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("hypercli/common")
}

func GetCmdClient(c *cli.Context) *CmdClient {
	client := NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
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
func (cc *CmdClient) InvokeCmd(cmd *admin.Command) *admin.CommandResult {
	logger.Critical(cmd.ToJson())
	rs, err := cc.Call(cmd.ToJson())
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return rs
}

func (cc *CmdClient) Call(cmd string) (*admin.CommandResult, error) {
	reqJson := []byte(cmd)
	urlStr := fmt.Sprintf("http://%s:%s", cc.host, cc.port)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	rs, err := cc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil, err
	}
	result := string(body)
	logger.Info(result)
	fmt.Printf(result)
	return nil, nil
}
