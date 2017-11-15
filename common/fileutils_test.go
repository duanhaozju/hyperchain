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
}
