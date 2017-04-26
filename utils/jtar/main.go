package main

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"io/ioutil"
	p "path"
	"strings"
	"os/exec"
	"hyperchain/common"
)

const NOT_EXIST = "not exist"

var source = flag.String("s","", "java contract directory path")
var target = flag.String("t","", "output path")

func main() {
	flag.Parse()
	if !validationCheck(*source) {
		fmt.Println("invalid contract")
	} else {
		compress(*source, *target)
		readToBytes(*target)
	}
}

func validationCheck(path string) bool {
	validation := false
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".properties") {
			if propertyCheck(p.Join(path, file.Name())) {
				validation = true
			}
		}
	}
	return validation
}

func propertyCheck(path string) bool {
	ps := properties.MustLoadFile(path, properties.UTF8)
	cName := ps.GetString("contract.name", NOT_EXIST)
	if cName == NOT_EXIST {
		fmt.Println("no [contract.name] property specified")
		return false
	}
	mClass := ps.GetString("main.class", NOT_EXIST)
	if mClass == NOT_EXIST {
		return false
	}
	prefix := ps.GetString("package.prefix", NOT_EXIST)
	if prefix == NOT_EXIST {
		fmt.Println("no [package.prefix] property specified")
		return false
	}
	return true
}

func compress(source, target string) {
	cp := exec.Command("tar", "-czf", target, source)
	if err := cp.Run(); err != nil {
		fmt.Println(err.Error())
	}
}

func readToBytes(target string) {
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		return
	}
	fmt.Println(common.Bytes2Hex(buf))
}
