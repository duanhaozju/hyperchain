package jvm

import (
	"io/ioutil"
	"strings"
	"path"
	"hyperchain/crypto"
	"bytes"
	"os"
	"fmt"
 	"archive/tar"
	"io"
	// "github.com/magiconair/properties"
)

var InvalidSourceCodeErr = "invalid source contract code"
var CompileSourceCodeErr = "compile source contract failed"
var SigSourceCodeErr = "generate signature for contract failed"


type ContractProperties struct {
	MainClass    string
	Prefix       string
	ContractName string
}

func decompression(buf []byte, namespace string) {
	r := bytes.NewReader(buf)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Println("alalalal")
		}
		if isProperty(hdr.Name) {
			fmt.Println(tr)
		}
	}
}

func staticCheck() bool {
	return true
}

func compile() error {
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

func getContractHome(namespace string) (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	suffix := fmt.Sprintf("namespaces/%s/contracts", namespace)
	return path.Join(cur, suffix), nil
}


func isProperty(fn string) bool {
	if strings.HasSuffix(fn, ".properties") {
		return true
	} else {
		return false
	}
}