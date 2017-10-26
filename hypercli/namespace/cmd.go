//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"

	"hyperchain/api/admin"
	cm "hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/hypercli/common"

	"github.com/urfave/cli"
)

// NewNamespaceCMD newNsId namespace related
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
			Action: newNsId,
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

// start helps start the namespace manager.
func start(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_startNsMgr",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// stop helps stop the namespace manager.
func stop(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_stopNsMgr",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// newNsId generates a new namespace ID using the given directory.
func newNsId(c *cli.Context) error {
	dir := c.String("configDir")
	if len(dir) == 0 {
		fmt.Println("Please specify namespace config dir using '-d' flag!")
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
		fmt.Printf("newNsId namespace id is: %s\n", name)
	} else {
		fmt.Println("invalid peerconfig")
	}

	return nil
}

// generateNamespaceId generates a unique ID using hostsName and timestamp.
func generateNamespaceId(hostNames string) string {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	hashString := hostNames + time.Now().String() + strconv.Itoa(rand.Intn(1000000))
	hash := kec256Hash.Hash(hashString).Hex()

	return "ns_" + hash[2:14]
}

// startNamespace helps start a namespace.
func startNamespace(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	if c.NArg() != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	cmd := &admin.Command{
		MethodName: "admin_startNamespace",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// stopNamespace helps stop a namespace.
func stopNamespace(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	if c.NArg() != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	cmd := &admin.Command{
		MethodName: "admin_stopNamespace",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// restartNamespace helps restart a namespace.
func restartNamespace(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	if c.NArg() != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	cmd := &admin.Command{
		MethodName: "admin_restartNamespace",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// registerNamespace helps register a namespace.
func registerNamespace(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	if c.NArg() != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	cmd := &admin.Command{
		MethodName: "admin_registerNamespace",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// deregisterNamespace helps deregister a namespace.
func deregisterNamespace(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	if c.NArg() != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	cmd := &admin.Command{
		MethodName: "admin_deregisterNamespace",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// listNamespaces lists all namespace names the given nodes participates in.
func listNamespaces(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_listNamespaces",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
