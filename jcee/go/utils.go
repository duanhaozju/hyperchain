package jvm

import (
	"os"
	"path"
)

const (
	ContractHome   = "hyperjvm/contracts"
	BinHome        = "hyperjvm/bin"
	ContractPrefix = "contract"
	CompressFileN  = "contract.tar.gz"
)


func getContractDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, ContractHome), nil
}

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}

func cd(target string, relative bool) (string, error)  {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if relative {
		err := os.Chdir(path.Join(cur, target))
		if err != nil {
			return "", err
		}
	} else {
		err := os.Chdir(target)
		if err != nil {
			return "", err
		}
	}
	return cur, nil
}
