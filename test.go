package main

import (
	"encoding/json"
	"fmt"
)

type Person struct{
	Name string
	Age int
}

func main() {

	var zhangsan = Person{Name:"zhansgsan",Age:10}
	zhangsans , _:= json.Marshal(zhangsan)
	fmt.Println(string(zhangsans))

	var i interface{}
	json.Unmarshal(zhangsans, &i)
	fmt.Println(i)
	lisi := i.(Person)
	fmt.Println("lisi: ", lisi)
}