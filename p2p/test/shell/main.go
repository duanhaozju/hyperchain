package main

import (
	"github.com/abiosoft/ishell"
	"strings"
)

func main() {
	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Println("Sample Interactive Shell")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "greet",
		Help: "greet user",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(true)
			c.Println("Hello", strings.Join(c.Args, " "))
		},
	})

	// run shell
	shell.Run()
}