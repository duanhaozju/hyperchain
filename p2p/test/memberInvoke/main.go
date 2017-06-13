package main

import "fmt"

type ABC struct {
	name string
}

func (abc *ABC) say() {
	fmt.Println(abc.name)
}

func (abc *ABC) Handle(foo func()) {
	foo()
}

func main() {
	abc := ABC{
		name: "zhangshan",
	}
	abc.Handle(abc.say)
}
