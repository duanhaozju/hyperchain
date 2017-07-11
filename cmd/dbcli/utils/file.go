package utils

import (
	"fmt"
	"os"
)

func CreateOrAppend(path string, content string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println(err.Error())
	}
}
