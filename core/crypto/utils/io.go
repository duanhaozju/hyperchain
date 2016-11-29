//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package utils

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
)

// DirMissingOrEmpty checks is a directory is missin or empty
func DirMissingOrEmpty(path string) (bool, error) {
	dirExists, err := DirExists(path)
	if err != nil {
		return false, err
	}
	if !dirExists {
		return true, nil
	}

	dirEmpty, err := DirEmpty(path)
	if err != nil {
		return false, err
	}
	if dirEmpty {
		return true, nil
	}
	return false, nil
}

// DirExists checks if a directory exists
func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DirEmpty checks if a directory is empty
func DirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// FileMissing checks if a file is missing
func FileMissing(path string, name string) (bool, error) {
	_, err := os.Stat(filepath.Join(path, name))
	if err != nil {
		return true, err
	}
	return false, nil
}

// FilePathMissing returns true if the path is missing, false otherwise.
func FilePathMissing(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return true, err
	}
	return false, nil
}

// DecodeBase64 decodes from Base64
func DecodeBase64(in string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(in)
}

// EncodeBase64 encodes to Base64
func EncodeBase64(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

// IntArrayEquals checks if the arrays of ints are the same
func IntArrayEquals(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
