package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
	Io utils
*/
// Read n entry from file specified by path
// Return actual entries number and entries content
func ReadEntries(path string, n int) (int, []string) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil
	}
	defer file.Close()
	var ret []string
	var count int = 0
	// TODO read from file randomly
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && count < n {
		ret = append(ret, scanner.Text())
		count += 1
	}
	if err := scanner.Err(); err != nil {
		return 0, nil
	}
	return len(ret), ret
}

// Write to file specified by path
func WriteEntries(path string, data []string) error {
	var file *os.File
	var err error
	idx := strings.LastIndex(path, "/")
	if idx != -1 {
		subpath := path[:idx]
		err := os.MkdirAll(subpath, 0777)
		if err != nil {
			return err
		}
	}
	if _, e := os.Stat(path); os.IsNotExist(e) {
		file, err = os.Create(path)
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		file, err = os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	}
	for _, entry := range data {
		if !strings.HasSuffix(entry, "\n") {
			entry = entry + "\n"
		}
		_, err = file.WriteString(entry)
	}
	file.Close()
	return nil
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}
	return
}

func CopyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourceFilePointer := source + "/" + obj.Name()
		destinationFilePointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourceFilePointer, destinationFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourceFilePointer, destinationFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

// Output to stdout
func Output(silense bool, msg ...interface{}) {
	if !silense {
		// TODO change logger Level
		fmt.Println(msg...)
	}
}
