package main

/*
#cgo CFLAGS : -I./include
#cgo LDFLAGS: -L./lib -lsydapi

#include "sydapi.h"
*/
import "C"
import ("fmt")

func main() {
	cs1 := C.CString("118.89.150.87")
	//C.free(unsafe.Pointer(cs))
	fmt.Println(cs1);
}
