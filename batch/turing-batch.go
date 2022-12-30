package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	//csvFileName = "test.csv"
	//csvFileName = "古诗词.csv"
	csvFileName         = "alpha-request.csv"
	resultFileName      = "result.csv"
	url                 = "http://api.turingos.cn/turingos/api/v2"
	requestJsonTemplate = `{
			"data":
			{
				"content": [{
					"data": "%s"
				}],
				"userInfo":
				{
					"uniqueId": "9727fcfaa7ee49589f3fe71af0358a31"
				}
			},
			"key": "9727fcfaa7ee49589f3fe71af0358a31",
			"timestamp": "1654170242136"
		}`
)

type JsonResult struct {
	GlobalId string `json:"globalId"`
	Intent   struct {
		Code         int         `json:"code"`
		OperateState int         `json:"operateState"`
		Parameters   interface{} `json:"parameters"`
	} `json:"intent"`
	Results []struct {
		GroupType  int    `json:"groupType"`
		ResultType string `json:"resultType"`
		Values     struct {
			Text string `json:"text"`
		} `json:"values"`
	} `json:"results"`
}

type CSVResult struct {
	respTime     string
	skillQuery   string
	query        string
	globalId     string
	code         string
	paramtersMsg string
	results      string
	text         string
	resultJson   string
}

func (r *CSVResult) lines() []string {
	fmt.Println("r: ", r)
	return []string{
		r.respTime,
		r.skillQuery,
		r.query,
		r.globalId,
		r.code,
		r.paramtersMsg,
		r.results,
		r.text,
		r.resultJson,
	}
}

func getFile(fileName string, flag int) *os.File {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Can not get workDir")
	}

	csvPath := path.Join(workDir, fileName)
	file, err := os.OpenFile(csvPath, flag, os.ModeType)

	if err != nil {
		log.Fatalln("OpenFile error: ", err)
	}

	return file
}

func readCsv(fileName string) (results []string) {

	file := getFile(fileName, os.O_RDONLY)
	defer func() {
		_ = file.Close()
	}()

	reader := csv.NewReader(file)
	for {
		read, err := reader.Read()
		if err == io.EOF {
			break
		}
		results = append(results, read[0])
	}

	if results != nil && len(results) > 1 {
		results = results[1:]
	}

	return results
}

func httpTuring(data string) *JsonResult {
	fmt.Println("请求图灵接口，查询词语为: ", data)
	requestJson := fmt.Sprintf(requestJsonTemplate, data)
	buffer := bytes.NewBuffer([]byte(requestJson))
	request, err := http.NewRequest(http.MethodPost, url, buffer)

	if err != nil {
		log.Fatalln("Request Error: ", err)
	}

	resp, err := http.DefaultClient.Do(request)
	defer func() {
		_ = resp.Body.Close()
	}()

	if err != nil {
		log.Fatalln("Request Error: ", err)
	}

	result := &JsonResult{}
	jsonBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln("Read Json Bytes Error: ", err)
	}

	err = json.Unmarshal(jsonBytes, result)

	if err != nil {
		fmt.Println("Parse Error: ", string(jsonBytes))
		log.Fatalln("Json Parse Error: ", err)
	}

	return result
}

func csvEncode(jsonResult *JsonResult) *CSVResult {
	parameters, _ := json.Marshal(jsonResult.Intent.Parameters)
	results, _ := json.Marshal(jsonResult.Results)
	texts := make([]string, 0)

	for _, result := range jsonResult.Results {
		text := strings.TrimSpace(result.Values.Text)
		texts = append(texts, text)
	}

	jsonR, _ := json.Marshal(jsonResult)
	result := &CSVResult{
		skillQuery:   "指尖查词",
		globalId:     jsonResult.GlobalId,
		code:         strconv.FormatInt(int64(jsonResult.Intent.Code), 10),
		paramtersMsg: string(parameters),
		results:      string(results),
		text:         strings.Join(texts, ","),
		resultJson:   string(jsonR),
	}

	return result
}

func RequestAPI(searchStr string) *CSVResult {
	startTime := time.Now()
	jsonResult := httpTuring(searchStr)
	endTime := time.Now()

	//将返回的json转换为 CSVResult
	csvResult := csvEncode(jsonResult)

	//request请求执行时间
	requestTime := fmt.Sprintf("%dms", endTime.Sub(startTime).Milliseconds())
	csvResult.respTime = requestTime

	//设置请求查询的字符串
	csvResult.query = searchStr

	return csvResult
}

//启动携程, 用于请求数据
func startGroup(buf <-chan string, result chan<- *CSVResult, cancelFunc context.CancelFunc, resultSize int) {

	var counter int64

	for i := 0; i < 16; i++ {
		go func() {
			for counter < int64(resultSize) {
				searchStr := <-buf
				csvResult := RequestAPI(searchStr)
				result <- csvResult
				atomic.AddInt64(&counter, 1)
			}

			//当携程全部执行完毕后关闭主线程阻塞
			if counter >= int64(resultSize) {
				cancelFunc()
			}

		}()
	}
}

func Start() {

	//从 csv 读取数据
	results := readCsv(csvFileName)
	fmt.Println(results)

	ctx, cancelFunc := context.WithCancel(context.Background())

	buf := make(chan string, 512)
	result := make(chan *CSVResult, 512)

	//启动携程
	startGroup(buf, result, cancelFunc, len(results))

	//启动携程
	for _, searchStr := range results {
		//向队列添加数据
		buf <- searchStr
	}

	//将请求结果写入到chan buffer
	writeResult(resultFileName, result, ctx)
}

//测试返回结果
func printResult(fileName string, csvChan <-chan *CSVResult, ctx context.Context) {

one:
	for {
		select {
		case result := <-csvChan:
			fmt.Println("CSV Result: ", result.lines())
		case <-ctx.Done():
			fmt.Println("执行完毕")
			break one
		}
	}

}

func writeResult(fileName string, csvChan <-chan *CSVResult, ctx context.Context) {

	//如果存在文件，则将当前文件删除
	stat, _ := os.Stat(fileName)
	if stat != nil {
		_ = os.Remove(fileName)
	}

	file := getFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND)

	defer func() {
		_ = file.Close()
	}()

	writer := csv.NewWriter(file)

	_ = writer.Write([]string{
		"resTime",
		"skill_query",
		"query",
		"globalId",
		"code",
		"parameters_msg",
		"results",
		"Text",
		"返回Json",
	})

one:
	for {
		select {
		case result := <-csvChan:
			fmt.Println("CSV Result: ", result.lines())
			err := writer.Write(result.lines())
			if err != nil {
				log.Fatalln("写入文件出错, 原因为: ", err)
			}
		case <-ctx.Done():
			fmt.Println("执行完毕")
			break one
		}
	}

	writer.Flush()
}
