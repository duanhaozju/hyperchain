package namespace

import (
	"os"
	"path"
)

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}

