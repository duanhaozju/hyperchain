//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"fmt"
	"github.com/urfave/cli"
	admin "hyperchain/api/admin"
	cm "hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/hypercli/common"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"
)

//NewNamespaceCMD new namespace related
func NewNamespaceCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "startNsMgr",
			Usage:  "start the namespace manager",
			Action: start,
		},
		{
			Name:   "stopNsMgr",
			Usage:  "stop the namespace manager",
			Action: stop,
		},
		{
			Name:   "new",
			Usage:  "generate a unique namespace id",
			Action: new,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "configDir, d",
					Value: "",
					Usage: "specify the namespace's config dir",
				},
			},
		},
		{
			Name:   "start",
			Usage:  "start namespace",
			Action: startNamespace,
		},
		{
			Name:   "stop",
			Usage:  "stop namespace",
			Action: stopNamespace,
		},
		{
			Name:   "restart",
			Usage:  "restart namespace",
			Action: restartNamespace,
		},
		{
			Name:   "register",
			Usage:  "register namespace",
			Action: registerNamespace,
		},
		{
			Name:   "deregister",
			Usage:  "deregister namespace",
			Action: deregisterNamespace,
		},
		{
			Name:   "list",
			Usage:  "list current registered namespaces",
			Action: listNamespaces,
		},
	}
}

func start(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startNsMgr",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func stop(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopNsMgr",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func new(c *cli.Context) error {
	dir := c.String("configDir")
	if len(dir) == 0 {
		fmt.Printf("%s is invalid namespace config dir!", dir)
		return nil
	}
	configPath := path.Join(dir, "peerconfig.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Can not find peerconfig.toml in config path: %s\n", dir)
		return nil
	}
	config := cm.NewConfig(configPath)
	nodes := config.Get("nodes")
	var hostNames string
	if nodemap, ok := nodes.([]interface{}); ok {
		for _, n := range nodemap {
			if node, ok := n.(map[string]interface{}); ok {
				//fmt.Printf("node name %s, node value %s \n", "hostname", node["hostname"])
				hostName, _ := node["hostname"].(string)
				hostNames += hostName
			} else {
				fmt.Println("invalid peerconfig")
				return nil
			}
		}
		name := generateNamespaceId(hostNames)
		fmt.Printf("new namespace id is: %s\n", name)
	} else {
		fmt.Println("invalid peerconfig")
	}

	return nil
}

//namespace generate functions
func generateNamespaceId(hostNames string) string {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	hashString := hostNames + time.Now().String() + strconv.Itoa(rand.Intn(1000000))
	//fmt.Println(hashString)
	hash := kec256Hash.Hash(hashString).Hex()
	//fmt.Println(hash)
	return "ns_" + hash[2:14]
}

func startNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func stopNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func restartNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func registerNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_registerNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func deregisterNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_deregisterNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func listNamespaces(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_listNamespaces",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
