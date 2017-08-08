//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"io"
	"path"
)

// MakeName creates a node name that follows the ethereum convention
// for such names. It adds the operation system name and Go runtime version
// the name.
func MakeName(name, version string) string {
	return fmt.Sprintf("%s/v%s/%s/%s", name, version, runtime.GOOS, runtime.Version())
}

func ExpandHomePath(p string) (path string) {
	path = p
	sep := string(os.PathSeparator)

	// Check in case of paths like "/something/~/something/"
	if len(p) > 1 && p[:1+len(sep)] == "~"+sep {
		usr, _ := user.Current()
		dir := usr.HomeDir

		path = strings.Replace(p, "~", dir, 1)
	}

	return
}

func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func AbsolutePath(Datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(Datadir, filename)
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}


func GetGoPath() string {
	env := os.Getenv("GOPATH")
	l := strings.Split(env, ":")
	if len(l) > 1 {
		return l[len(l) - 1]
	}
	return l[0]
}

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

// GetPath get complete path for namespace level config file.
func GetPath(namespace, shortPath string) string {
	if len(namespace) == 0 {
		return shortPath
	}
	return path.Join("namespaces", namespace, shortPath)
}
