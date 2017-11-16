//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"bufio"
)

func TestFileExist(t *testing.T) {
	assert.Equal(t, true, FileExist(path.Join(testResourcePrefix, "global.toml")), "This file should exist")
	assert.Equal(t, false, FileExist(path.Join(testResourcePrefix, "gl.toml")), "This file should not exist")
}

func TestGetPath(t *testing.T) {
	namespace := "global"
	shortPath := "something"
	assert.Equal(t, "namespaces/"+namespace+"/"+shortPath, GetPath(namespace, shortPath), "should be equal")
}

func TestSeekAndAppend(t *testing.T) {
	filePath := "./testhelper/resources/namespace.toml"
	err := SeekAndAppend("[consensus.rbft]", filePath, fmt.Sprintf("    %s", "test = 100"))
	if err != nil {
		t.Error(err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, strings.Contains(string(data), "test = 100"), "The config file should contain this message")

	err = SeekAndDelete("[consensus.rbft]", filePath)
	if err != nil {
		t.Error(err)
	}
	data, err = ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, false, strings.Contains(string(data), "test = 100"), "The config file should not contain this message")
}

func SeekAndDelete(item, filePath string) error {

	//1. check validity of arguments.
	if len(filePath) == 0 {
		return fmt.Errorf("Invalid filePath, file path is empty!")
	}

	if len(item) == 0 {
		return fmt.Errorf("Invalid iterm, item is empty")
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

	scanned := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !scanned {
			if _, err = newFile.WriteString(line + "\n"); err != nil {
				return err
			}
		} else {
			scanned = false
		}

		if strings.Contains(line, item) {
			scanned = true
		}
	}

	//4. rename .tmp file to original file name
	if err = os.Rename(newFilePath, filePath); err != nil {
		return err
	}
	return nil
}
