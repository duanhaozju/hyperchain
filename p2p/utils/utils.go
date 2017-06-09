package utils

import (
	"os"
	"strings"
	"path/filepath"
)

func GetProjectPath() string{
	gopath := os.Getenv("GOPATH")
	if strings.Contains(gopath,":"){
		gopath = strings.Split(gopath,":")[0]
	}
	return gopath
}

func GetCurrentDirectory() (string,error){
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "",err
	}
	return strings.Replace(dir, "\\", "/", -1),nil

}
