package jc_maker

import (
	"flag"
	"path/filepath"
	"os"
)

var path = flag.String("p", "", "java contract directory path")

func main() {
	flag.Parse()
}



func propertyCheck(path string) bool {
	filepath.Walk(path, visitProperty)
	return true
}

func visitProperty(path string, info os.FileInfo, err error) error {

}



