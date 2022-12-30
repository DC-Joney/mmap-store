package store

import (
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"path/filepath"
	"strings"
	"time"
	"turing/resolve/statics"
)

var (
	poolSize                         = 1 << 3
	allocateService *AllocateService = &AllocateService{}
)

type AllocateRequest struct {
	FileName   string
	stopSh     chan struct{}
	mappedFile *MappedFile
	fileSize   int64
}

func (req AllocateRequest) Done() <-chan struct{} {
	return req.stopSh
}

func (req AllocateRequest) Stop() {
	req.stopSh <- struct{}{}
	close(req.stopSh)
}

func NewAllocateRequest(fileName string) *AllocateRequest {
	return &AllocateRequest{
		FileName: fileName,
		stopSh:   make(chan struct{}),
	}
}

type AllocateService struct {
	//用于存储创建好的MappedFile文件
	requestMap map[string]*AllocateRequest
	Pool       *ants.PoolWithFunc
}

func (service AllocateService) AddRequest(nextFile, nextNextFile string) (*MappedFile, error) {
	if strings.TrimSpace(nextFile) == "" {
		return nil, errors.New("nextFile name must not be null")
	}
	if strings.TrimSpace(nextNextFile) == "" {
		return nil, errors.New("nextNextFile name must not be null")
	}

	//判断数据是否已经存在
	request, ok := service.requestMap[nextFile]
	if !ok {
		request = NewAllocateRequest(nextFile)
		service.requestMap[nextFile] = request
		//添加请求
		_ = service.Pool.Invoke(request)
	}

	//判断下下个文件是否也已经创建了
	_, ok = service.requestMap[nextNextFile]
	if !ok {
		nextRequest := NewAllocateRequest(nextNextFile)
		//添加请求
		_ = service.Pool.Invoke(nextRequest)
	}

	statics.Logger.Info(nextFile)
	statics.Logger.Info(nextNextFile)

	timer := time.NewTimer(time.Second * 5)
	timeout := false

	select {
	//等待请求创建完成
	case <-request.Done():
	//超时时间为5s
	case <-timer.C:
		timeout = true
	}

	if timeout {
		return nil, errors.New("Get file timeout")
	}

	result := service.requestMap[nextFile]

	//删除数据
	delete(service.requestMap, nextFile)
	return result.mappedFile, nil
}

// Start 启动内部携程
func (service *AllocateService) Start() {
	handler := ants.WithPanicHandler(func(i interface{}) {
		statics.Logger.Error(i)
	})
	service.requestMap = make(map[string]*AllocateRequest)
	service.Pool, _ = ants.NewPoolWithFunc(poolSize, service.createFile, handler)
}

// createFile 创建文件的流程
func (service *AllocateService) createFile(data interface{}) {
	request := data.(*AllocateRequest)
	fileName := request.FileName
	statics.Logger.Infof("接收到创建请求: %s", request)
	mappedFile := NewMappedFile(fileName, fileSize, false)
	request.mappedFile = mappedFile
	//创建完成后不在阻塞创建线程
	request.Stop()
}

type MappedFileQueue struct {
	//文件目录
	FileDir string
	//目录下的所有mmapFile文件
	mappedFiles []*MappedFile
	//flush的位置，对于所有的文件而言
	flushWhere int64
	//每个文件的大小
	FileSize int64
}

// GetLastMappedFile :获取最后一个文件
// needCreate: 当没有文件时是否需要创建
func (this *MappedFileQueue) GetLastMappedFile(needCreate bool) *MappedFile {

	var createOffset int64 = -1
	fileLast := this.getLastFile()
	if fileLast == nil {
		createOffset = 0
	}

	if fileLast != nil && fileLast.IsFull() {
		createOffset = fileLast.fileFromOffset + this.FileSize
	}

	if createOffset != -1 && needCreate {
		fileName := fmt.Sprintf("%020d", createOffset)
		nextFileName := fmt.Sprintf("%020d", createOffset+this.FileSize)

		//拼接下一个文件的的路径
		nextFile := filepath.Join(this.FileDir, fileName)
		nextNextFile := filepath.Join(this.FileDir, nextFileName)

		mappedFile, err := allocateService.AddRequest(nextFile, nextNextFile)
		if err != nil {
			statics.Logger.Error("Create MappedFile error: ", err)
		}

		this.mappedFiles = append(this.mappedFiles, mappedFile)
		return mappedFile
	}

	return fileLast
}

// getLastFile 获取最新的 MappedFile 文件
func (this MappedFileQueue) getLastFile() *MappedFile {
	fileCount := len(this.mappedFiles)
	if fileCount > 0 {
		return this.mappedFiles[fileCount-1]
	}

	return nil
}

func init() {
	allocateService.Start()
}
