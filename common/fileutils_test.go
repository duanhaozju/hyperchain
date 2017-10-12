//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestFileExist(t *testing.T) {
	directoryPath := os.Getenv("GOPATH") + "/src/hyperchain/common/testhelper/resources"
	assert.Equal(t, true, FileExist(directoryPath+"/global.toml"), "This file should exist")

	assert.Equal(t, false, FileExist(directoryPath+"/gl.toml"), "This file should not exist")
}

func TestGetPath(t *testing.T) {
	namespace := "global"
	shortPath := "something"
	assert.Equal(t, "namespaces/"+namespace+"/"+shortPath, GetPath(namespace, shortPath), "should be equal")
}

func TestSeekAndAppend(t *testing.T) {
	filePath := "./testhelper/resources/global.toml"
	err := SeekAndAppend("[namespace.start]", filePath, fmt.Sprintf("    %s", "ns1 = true"))
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

	assert.Equal(t, true, strings.Contains(string(data), "ns1 = true"), "The config file should contain this message")
}
