package common

import (
	"os"
	"path/filepath"
)

func GetExeDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir, err := filepath.Abs(filepath.Dir(ex))
	if err != nil {
		return "", err
	}
	return dir, nil
}
