package jvm

import (
	"io/ioutil"
	"strings"
	"path"
	"bytes"
	command "os/exec"
	"path/filepath"
	"os"
	"fmt"
)

var DecompressErr = "decompress source contract failed"
var InvalidSourceCodeErr = "invalid source contract code"
var CompileSourceCodeErr = "compile source contract failed"
var SigSourceCodeErr = "generate signature for contract failed"


type ContractProperties struct {
	MainClass    string
	Prefix       string
	ContractName string
}

func decompression(buf []byte) (string, error) {
	hPath, err := getContractDir()
	if err != nil {
		return "", err
	}
	tmpDir, err := ioutil.TempDir(hPath, ContractPrefix)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(path.Join(tmpDir, CompressFileN), buf, 0644)
	if err != nil {
		return "", err
	}

	cmd := command.Command("tar", "-xzf", path.Join(tmpDir, CompressFileN), "-C", path.Join(tmpDir))
	if err = cmd.Run(); err != nil {
		return "", err
	}

	if err = extractProperty(tmpDir); err != nil {
		return "", err
	}
	if err = moveFromTarDir(tmpDir); err != nil{
		return "",err
	}

	return tmpDir, nil
}

func staticCheck() bool {
	return true
}

func compile(contractPath string) error {
	binHome, err := getBinDir()
	if err != nil {
		return err
	}
	cmdPath := path.Join(binHome, "contract_compile.sh")
	cmd := command.Command(cmdPath, contractPath, contractPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func combineCode(dirPath string) ([]byte, error) {
	var buffer bytes.Buffer
	err := filepath.Walk(dirPath,
		func(p string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if !f.IsDir() && strings.HasSuffix(f.Name(), "class") {
				buf, err := ioutil.ReadFile(p)
				if err != nil {
					return err
				}
				buffer.Write(buf)
			}
			return nil
		})
	if err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}





func isProperty(fn string) bool {
	if strings.HasSuffix(fn, ".properties") {
		return true
	} else {
		return false
	}
}

func extractProperty(dirPath string) error {
	return filepath.Walk(dirPath,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if !f.IsDir() && (strings.HasSuffix(f.Name(), "properties") || strings.HasSuffix(f.Name(), "yaml") ) {
				cmd := command.Command("mv", path, dirPath)

				if err := cmd.Run(); err != nil {
					return err
				}
			}
			return nil
		})
}
func moveFromTarDir(dirPath string) error {
	dir_list,_ := ioutil.ReadDir(dirPath)
	var tmpDir string


	for _,dir := range dir_list{
		tmpDir = dir.Name()
		err := os.Rename(dirPath + "/" + tmpDir, dirPath)
		fmt.Println(err)
	}
	//cmd := command.Command("mv",dirPath+"/"+tmpDir+"/*",dirPath)
	//fmt.Println(dirPath+"/"+tmpDir+"/*")
	//if err := cmd.Run(); err != nil {
	//	return err
	//}
	return nil
}
