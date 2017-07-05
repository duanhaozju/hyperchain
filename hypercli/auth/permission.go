package auth

import (
	"github.com/urfave/cli"
	admin "hyperchain/api/jsonrpc/core"

	"fmt"
	"hyperchain/hypercli/common"
	"os"
	"sort"
)

// createUser implements the create user logic in hypercli, this method can only be called by root successfully
func createUser(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_createUser",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 2 {
		fmt.Println("Need 2 params(username and password)")
		return common.ErrInvalidArgsNum
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// alterUser implements the alter user logic in hypercli, this method can only be called by root successfully
func alterUser(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_alterUser",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 2 {
		fmt.Println("Need 2 params(username and new password)")
		return common.ErrInvalidArgsNum
	}
	result:= client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// dropUser implements the drop user logic in hypercli, this method can only be called by root successfully
func dropUser(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_delUser",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println("Need only 1 param(username)")
		return common.ErrInvalidArgsNum
	}
	result:= client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// grant grants the permissions to given username
func grant(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var permissions []string
	if c.String("path") != "" {
		args := c.Args()
		if len(args) != 1 {
			fmt.Println("Need only 1 param(username) because you have specify the permission file path")
			return common.ErrInvalidArgsNum
		}
		// get username
		permissions = append(permissions, args[0])

		// get permissions
		pms, err := common.ReadPermissionsFromFile(c.String("path"))
		permissions = append(permissions, pms...)
		if err != nil {
			fmt.Println("Read file failed: ", err)
			os.Exit(1)
		}
	} else {
		permissions = c.Args()
		if len(permissions) < 2 {
			fmt.Println("Need at least 2 params(username and permissions)")
			return common.ErrInvalidArgsNum
		}
		if len(permissions) == 2 {
			permissions = append(permissions, "all")
		}
	}

	cmd := &admin.Command{
		MethodName: "admin_grantPermission",
		Args:       permissions,
	}

	result:= client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// revoke revokes the permissions from given username
func revoke(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var permissions []string
	if c.String("path") != "" {
		args := c.Args()
		if len(args) != 1 {
			fmt.Println("Need only 1 param(username) because you have specify the permission file path")
			return common.ErrInvalidArgsNum
		}
		// get username
		permissions = append(permissions, args[0])

		// get permissions
		pms, err := common.ReadPermissionsFromFile(c.String("path"))
		permissions = append(permissions, pms...)
		if err != nil {
			fmt.Println("Read file failed: ", err)
			os.Exit(1)
		}
	} else {
		permissions = c.Args()
		if len(permissions) < 2 {
			fmt.Println("Need at least 2 params(username and permissions)")
			return common.ErrInvalidArgsNum
		}
		if len(permissions) == 2 {
			permissions = append(permissions, "all")
		}
	}

	cmd := &admin.Command{
		MethodName: "admin_revokePermission",
		Args:       permissions,
	}

	result:= client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// list lists the permission of current user
func list(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	username := c.Args()
	if len(username) == 0 {
		userinfo := new(common.UserInfo)
		err := common.ReadFile(tokenpath, userinfo)
		if err == nil {
			username = []string{userinfo.Username}
		}
	}
	if len(username) != 1 {
		fmt.Println("Need only 1 params(username)")
		return common.ErrInvalidArgsNum
	}
	cmd := &admin.Command{
		MethodName: "admin_listPermission",
		Args:       username,
	}

	result:= client.InvokeCmd(cmd)

	response, err := common.GetJSONResponse(result)
	if err != nil {
		fmt.Println("Get json response failed: ", err)
		os.Exit(1)
	}

	if permissions, ok := response.Result.([]interface{}); !ok {
		fmt.Printf("Sorry, %s have no permissions to do anything, " +
			"please contact to the administrator...\n", username)
		return nil
	} else {
		if len(permissions) == 0 {
			fmt.Printf("Sorry, %s have no permissions to do anything, " +
				"please contact to the administrator...\n", username)
			return nil
		}
		fmt.Printf("%s has permissions to: \n", username)
		var readablePermissions []string
		for _, permission := range permissions {
			readablePermissions = append(readablePermissions, admin.ReadablePermission(permission.(float64)))
		}
		sort.Strings(readablePermissions)
		for _, readablePermission := range readablePermissions {
			fmt.Println(readablePermission)
		}
	}

	return nil
}
