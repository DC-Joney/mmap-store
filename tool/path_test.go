package tool

import (
	"fmt"
	"log"
	"testing"
)

func TestGetUserDir(t *testing.T) {

	dir, err := GetResourceDir()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(dir)

}
