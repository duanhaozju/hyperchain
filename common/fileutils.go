//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// FileExist checks a file exists or not
func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// GetGoPath gets the GOPATH in this environment
func GetGoPath() string {
	env := os.Getenv("GOPATH")
	l := strings.Split(env, ":")
	if len(l) > 1 {
		return l[len(l)-1]
	}
	return l[0]
}

// GetPath get complete path for namespace level file
func GetPath(namespace, shortPath string) string {
	if len(namespace) == 0 {
		return shortPath
	}
	return path.Join("namespaces", namespace, shortPath)
}

// SeekAndAppend seek item by pattern in file whose path is filepath and than append content after that item
func SeekAndAppend(item, filePath, appendContent string) error {

	//1. check validity of arguments.
	if len(filePath) == 0 {
		return fmt.Errorf("Invalid filePath, file path is empty!")
	}

	if len(item) == 0 {
		return fmt.Errorf("Invalid iterm, item is empty")
	}

	if len(appendContent) == 0 {
		return fmt.Errorf("Invalid iterm, appenContent is empty")
	}

	//2. create a temp file to store new file content.

	newFilePath := filePath + ".tmp"

	newFile, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	defer newFile.Close()

	//3. copy and append appendContent

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err = newFile.WriteString(line + "\n"); err != nil {
			return err
		}
		if strings.Contains(line, item) {
			if _, err = newFile.WriteString(appendContent + "\n"); err != nil {
				return err
			}
		}
	}

	//4. rename .tmp file to original file name
	if err = os.Rename(newFilePath, filePath); err != nil {
		return err
	}
	return nil
}
