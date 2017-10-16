package jvm

import (
	"bytes"
	"io/ioutil"
	"os"
	command "os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var (
	DecompressErr        = "decompress source contract failed"
	InvalidSourceCodeErr = "invalid source contract code"
	CompileSourceCodeErr = "compile source contract failed"
	SigSourceCodeErr     = "generate signature for contract failed"
)

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

	cmd := command.Command("tar", "-xf", path.Join(tmpDir, CompressFileN), "-C", path.Join(tmpDir), "--strip-components=1")
	if err = cmd.Run(); err != nil {
		return "", err
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
	classData := make(map[string][]byte)
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
				classData[f.Name()] = buf
				//buffer.Write(buf)
			}
			return nil
		})
	if err != nil {
		return nil, err
	} else {
		var keys []string
		for k := range classData {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			buffer.Write(classData[k])
		}
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
