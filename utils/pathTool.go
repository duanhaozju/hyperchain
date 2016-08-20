package utils

import (
	"os"
)

func GetBasePath() string {
	path, _ := os.Getwd()
	//path := os.TempDir()
	return path
}