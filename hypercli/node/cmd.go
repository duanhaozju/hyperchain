//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package node

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/hypercli/common"
	"hyperchain/rpc"
	"strconv"
)

//NewNodeCMD new node related commands.
func NewNodeCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "delete",
			Usage:  "delete a node from specified namespace",
			Action: delNode,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "setting the namespace to delete node from",
				},
				cli.StringFlag{
					Name:  "ip, i",
					Value: "127.0.0.1",
					Usage: "setting the host ip to delete node from",
				},
				cli.StringFlag{
					Name:  "port, p",
					Value: "8085",
					Usage: "setting the host port to delete node from",
				},
			},
		},
	}
}

type peerinfos struct {
	ips   []string
	ports []string
}

func delNode(c *cli.Context) error {
	namespace := common.GetNonEmptyValueByName(c, "namespace")
	ip := common.GetNonEmptyValueByName(c, "ip")
	port := common.GetNonEmptyValueByName(c, "port")

	nodehash, err := getDelNodeHash(namespace, ip, port)
	if err != nil {
		fmt.Println("Failed to get node hash, exit del node...")
		fmt.Println(err.Error())
		return err
	}

	peers, err := getPeerInfo(namespace, ip, port)
	if err != nil {
		fmt.Println("Failed to get peers info, exit del node...")
		fmt.Println(err.Error())
		return err
	}

	err = sendDelNode(namespace, nodehash, peers)
	if err != nil {
		fmt.Println("Failed to send delete node, exit del node...")
		fmt.Println(err.Error())
		return err
	}

	return nil

}

func getHttpResponse(namespace, ip, port, method, params string) (jsonrpc.JSONResponse, error) {
	var response jsonrpc.JSONResponse
	client := common.NewRpcClient(ip, port)
	cmd := fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, params)

	result, err := client.Call(cmd, method)
	if err != nil {
		return response, err
	}

	return common.GetJSONResponse(result)
}

func getDelNodeHash(namespace, ip, port string) (string, error) {
	response, err := getHttpResponse(namespace, ip, port, "node_getNodeHash", "[{}]")
	if err != nil {
		return "", err
	}

	if hash, ok := response.Result.(string); !ok {
		return "", fmt.Errorf("rpc result: %v can't parse to string", response.Result)
	} else {
		return hash, nil
	}
}

func getPeerInfo(namespace, ip, port string) (peerinfos, error) {
	var info peerinfos

	response, err := getHttpResponse(namespace, ip, port, "node_getNodes", "[{}]")
	if err != nil {
		return info, err
	}

	if result, ok := response.Result.([]interface{}); !ok {
		return info, fmt.Errorf("rpc result: %v can't parse to PeerInfos", response.Result)
	} else {
		for _, v := range result {
			for key, value := range v.(map[string]interface{}) {
				if key == "ip" {
					info.ips = append(info.ips, value.(string))
				}
				if key == "rpcport" {
					info.ports = append(info.ports, strconv.Itoa(int(value.(float64))))
				}
			}
		}
		return info, nil
	}

}

func sendDelNode(namespace, hash string, peers peerinfos) error {
	params := fmt.Sprintf("[{\"nodehash\":\"%s\"}]", hash)
	for i, ip := range peers.ips {
		fmt.Printf("send del node to %v:%v\n", ip, peers.ports[i])
		_, err := getHttpResponse(namespace, ip, peers.ports[i], "node_deleteVP", params)
		if err != nil {
			return err
		}
	}
	return nil
}
