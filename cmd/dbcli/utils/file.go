package utils

import (
	"fmt"
	"os"
)

func CreateOrAppend(path string, content string) *os.File {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println(err.Error())
	}
	return file
}

func Append(file *os.File, content string) {
	_, err := file.WriteString(content + "\n")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func Close(file *os.File) {
	if file != nil {
		err := file.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
