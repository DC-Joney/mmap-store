package config

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestConfig1(t *testing.T)  {

	fs := os.DirFS("..")

	fmt.Printf("%T",fs)

	dir := path.Dir(".")
	fmt.Println(dir)

	homeDir, _ := os.UserHomeDir()
	fmt.Println(homeDir)
}

func TestInitConfig(t *testing.T) {
	InitConfig()
	fmt.Printf("%#v", *datasource)
}
