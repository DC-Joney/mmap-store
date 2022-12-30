package store

import (
	"encoding/binary"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"k8s.io/apimachinery/pkg/util/errors"
	"os"
	"sync"
	"sync/atomic"
	"turing/resolve/statics"
)

var (
	//fileSize 文件大小为1M
	fileSize int64 = 1 << 20
)

type MappedFile struct {
	File *os.File

	FileName string

	//文件大小
	FileSize int64

	mmapRegion *mmap.MMap

	//刷盘位置
	flushPosition int64

	//写入磁盘位置
	writePosition int64

	//读写锁
	rwLock *sync.RWMutex

	//文件写入的起始位置
	fileFromOffset int64
}

func (this *MappedFile) Write(offset int64, bytes []byte) {
	writeLen := len(bytes)
	region := *this.mmapRegion
	copy(region[offset:int(offset)+len(bytes)], bytes)

	//计算写如长度
	atomic.AddInt64(&this.writePosition, int64(writeLen))
}

func (this *MappedFile) Append(bytes []byte) {
	this.Write(this.writePosition, bytes)
}

func (this *MappedFile) AppendString(dataStr string) {
	this.WriteString(this.writePosition, dataStr)
}

func (this *MappedFile) WriteString(offset int64, dataStr string) {
	this.Write(offset, []byte(dataStr))
}

func (this *MappedFile) PutInt64(offset int, i int64) {
	binary.BigEndian.PutUint64((*this.mmapRegion)[offset:offset+8], uint64(i))
	this.writePosition += 6
}

func (this *MappedFile) PutInt32(offset int, i int32) {
	binary.BigEndian.PutUint32((*this.mmapRegion)[offset:offset+4], uint32(i))
	this.writePosition += 4
}

func (this *MappedFile) PutInt16(offset int, i int16) {
	binary.BigEndian.PutUint16((*this.mmapRegion)[offset:offset+2], uint16(i))
	this.writePosition += 2
}

func (this *MappedFile) AppendInt64(i int64) {
	writePos := this.writePosition
	this.PutInt64(int(writePos), i)
}

func (this *MappedFile) AppendInt32(i int32) {
	writePos := this.writePosition
	this.PutInt32(int(writePos), i)
}

func (this *MappedFile) AppendInt16(i int16) {
	writePos := this.writePosition
	this.PutInt16(int(writePos), i)
}

func (this *MappedFile) Flush() {
	err := this.mmapRegion.Flush()
	if err != nil {
		statics.Logger.Error(err)
	}

	//计算写如长度
	atomic.CompareAndSwapInt64(&this.flushPosition, this.flushPosition, this.writePosition)
}

func (this MappedFile) String() string {
	return fmt.Sprintf("%s", *this.mmapRegion)
}

func (this MappedFile) Close() error {

	compositeError := make([]error, 0)
	err := this.mmapRegion.Flush()

	if err != nil {
		compositeError = append(compositeError, err)
	} else {
		//计算写如长度
		atomic.CompareAndSwapInt64(&this.flushPosition, this.flushPosition, this.writePosition)
	}
	unmapError := this.mmapRegion.Unmap()
	if unmapError != nil {
		compositeError = append(compositeError, err)
	}

	return errors.NewAggregate(compositeError)
}

func (this MappedFile) IsFull() bool {
	return this.writePosition == this.FileSize
}

func NewMappedFile(fileName string, fileSize int64, deleteIfExists bool) *MappedFile {
	file, err := openOrCreateFile(fileName, deleteIfExists)
	if err != nil {
		statics.Logger.Error("Get file err: ", err)
	}
	statics.Logger.Info("创建MappedFile开始")
	mappedRegion, err := mmap.MapRegion(file, int(fileSize), mmap.RDWR, 0, 0)

	if err != nil {
		statics.Logger.Error("Create mapped buffer error: ", err)
	}

	mappedFile := &MappedFile{
		mmapRegion: &mappedRegion,
		FileName:   fileName,
		FileSize:   fileSize,
		File:       file,
	}

	//stat, err := file.Stat()
	//size := stat.Size()
	//重置当前文件的writePosition 与 writePosition
	mappedFile.writePosition = 0
	mappedFile.flushPosition = mappedFile.writePosition
	statics.Logger.Info("创建MappedFile结束")
	return mappedFile
}

func openOrCreateFile(fileName string, deleteIfExists bool) (*os.File, error) {
	var file *os.File
	if deleteIfExists {
		file, err := os.Create(fileName)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	_, err := os.Stat(fileName)

	if err != nil {
		if _, ok := err.(*os.PathError); ok || err == os.ErrNotExist {
			statics.Logger.Info("创建新的文件")
			file, err = os.Create(fileName)
			if err != nil {
				return nil, err
			}
		}
	} else {
		file, err = os.OpenFile(fileName, os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}
