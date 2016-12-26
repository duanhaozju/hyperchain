package main
/*
#cgo CFLAGS : -I./include
#cgo LDFLAGS: -L./lib -ladd

#include "add.h"
*/
import "C"

func main() {
	fmt.Println("vim-go")
}
