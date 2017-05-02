package main

import (
	"fmt"
)

func PrintDefault() {
	fmt.Println("SYNOPSIS\n")
	// creation
	fmt.Println("Create jar file")
	fmt.Println("\tjtar cf jarfile [-C dir] inputdirs")
	fmt.Println()

	fmt.Println("where: \n")
	fmt.Println("cuxtf\n\tOptions that control the `jtar` command.")
	fmt.Println("jarfile")
	fmt.Println("\tFile name of the Jar file to be created (c), updated (u), extracted (x), or have its table of contents viewed (t). The -f option and filename jarfile are a pair -- if either is present, they must both appear. Note that omitting -f and jarfile accepts jar file from standard input (for x and t) or sends jar file to standard output (for c and u).")
}
