package jvm

import (
	"io/ioutil"
	"strings"
	"path"
	"hyperchain/crypto"
	"bytes"
	command "os/exec"
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
	oldPath, err := cd(binHome, false)
	if err != nil {
		return err
	}
	defer cd(oldPath, false)

	cmd := command.Command("./contract_compile.sh", contractPath, contractPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func signature(dirPath string) ([]byte, error) {
	var buffer bytes.Buffer
	files, err := ioutil.ReadDir(dirPath)

	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "class") {
			buf, err := ioutil.ReadFile(path.Join(dirPath, file.Name()))
			if err != nil {
				return nil, err
			}
			buffer.Write(buf)
		}
	}
	sig := crypto.Keccak256(buffer.Bytes())
	return sig, nil
}





func isProperty(fn string) bool {
	if strings.HasSuffix(fn, ".properties") {
		return true
	} else {
		return false
	}
}