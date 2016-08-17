package utils

import (
	"os"
)

func GetBasePath() string {
	path, _ := os.Getwd()
	return path
}