package ipc

import "github.com/abiosoft/ishell"

func IPCShell(endpoint string) {
	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Println("Welcome to Hypernet interactive shell!")

	// register a function for "greet" command.
	for _, cmd := range getCmds(endpoint) {
		shell.AddCmd(cmd)
	}

	shell.AutoHelp(true)
	shell.Println(shell.HelpText())
	// run shell
	shell.Run()
}
