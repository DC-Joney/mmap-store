package tool

import (
	"os"
	"path/filepath"
)

const (
	resourceName = "resources"
)

// GetResourceDir 获取resource文件目录
func GetResourceDir() (string, error) {
	userDir, err := GetUserDir()

	if err != nil {
		return "", nil
	}

	resourceDir := filepath.Join(userDir, resourceName)
	return resourceDir, nil
}

// GetUserDir 获取程序根目录
func GetUserDir() (string, error) {

	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	parentDir, _ := filepath.Split(currentPath)
	return parentDir, nil
}
