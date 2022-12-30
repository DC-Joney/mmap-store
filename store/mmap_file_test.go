package store

import (
	"fmt"
	"testing"
)

func TestNewMmapFile(t *testing.T) {
	mmapFile := NewMappedFile("test.txt", fileSize, false)
	mmapFile.Write([]byte("1234==========================================="))
	mmapFile.Flush()

	fmt.Printf("%s", mmapFile.mmapRegion)
}

func TestFormmater(t *testing.T) {
	sprintf := fmt.Sprintf("%020d", 3)
	fmt.Println(sprintf)
}

func TestGetLastFile(t *testing.T) {
	queue := MappedFileQueue{
		FileSize: fileSize,
		FileDir:  "D:\\go-project\\turing\\store",
	}

	mappedFile := queue.GetLastMappedFile(true)
	mappedFile.Write([]byte("12345"))
	_ = mappedFile.Close()
}
