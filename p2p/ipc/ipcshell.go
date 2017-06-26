package ipc

import "github.com/abiosoft/ishell"

func IPCShell(){
	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Println("Sample Interactive Shell")

	// register a function for "greet" command.
	for _,cmd := range getCmds(){
		shell.AddCmd(cmd)
	}
	// run shell
	shell.Run()
}
