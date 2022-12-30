package statics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BookPage struct {

	//内页图片地址
	ImageUrl string `json:"imgUrl"`

	//页面编号
	PageNo int `json:"pageNo"`

	Custom bool `json:"custom"`

	BookSign string `json:"bookSign,omitempty"`

	ApiKey string `json:"apiKey,omitempty"`

	Audios []struct {
		//对应的页面
		PageNo int `json:"pageNo"`

		//音频文本
		AudioText string `json:"audioText"`
	} `json:"bookPageMultiplyAudioDTOList"`
}

// 	{"apiKey":"9f6fdcedf5c44111b5e14d632c043189","bookSign":"ac4f89488fbd443683cf0e9cd151652f","pageSize":6,"pageNo":1}
type BookPageParam struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
	BookSign string `json:"bookSign"`
	ApiKey   string `json:"apiKey"`
}


type BookPageRequest struct {
	Pool     *ants.PoolWithFunc
	Receiver chan *BookPage
}

//InvokeRequest 请求Store库
func (request *BookPageRequest) InvokeRequest(param *BookPageParam) {
	err := request.Pool.Invoke(param)
	if err != nil {
		Logger.Error("Add BookPageParam to pools error: ", err)
	}
}


//requestStore 请求Store库
func (request *BookPageRequest) workerRequest(data interface{}) {
	defer func() {
		if err:= recover(); err != nil {
			Logger.Error("BookPage worker pool error: ", err)
		}
	}()

	param := data.(*BookPageParam)
	bookPages, err := RequestBookPage(param)
	if err != nil {
		Logger.Info("workFunc==> 请求BookPage出错: ", err)
		return
	}

	for _, bookPage := range bookPages {
		bookPage.BookSign = param.BookSign
		bookPage.ApiKey = param.ApiKey

		//统计当前数据
		Statics.AddBookPage(&bookPage)

		//todo 暂时不需要放入内页数据
		//将内页的数据数据放在队列中
		//request.Receiver <- &bookPage
	}
}

func NewBookPageRequest(poolSize int) *BookPageRequest {
	request := &BookPageRequest{
		Receiver: make(chan *BookPage, 1<<9),
	}

	request.Pool, _ = ants.NewPoolWithFunc(poolSize, request.workerRequest)
	return request
}

func NewBookPageParam(count int, bookSign, apiKey string) *BookPageParam {
	return &BookPageParam{
		PageNo:   1,
		PageSize: count + 1,
		BookSign: bookSign,
		ApiKey:   apiKey,
	}
}


//RequestBookPage 请求绘本内页列表
// 	{"apiKey":"9f6fdcedf5c44111b5e14d632c043189","bookSign":"ac4f89488fbd443683cf0e9cd151652f","pageSize":6,"pageNo":1}
func RequestBookPage(pageParam *BookPageParam) ([]BookPage, error) {
	pageUrl, err := url.JoinPath(BookIot.BaseURL, BookIot.BookPageList.Path)
	if err != nil {
		Logger.Error("BookPage join path error: ", err)
	}
	method := strings.ToUpper(BookIot.BookPageList.Method)
	requestJson, err := json.Marshal(pageParam)
	request, err := http.NewRequest(method, pageUrl, bytes.NewBuffer(requestJson))

	if err != nil {
		Logger.Error("BookPage request error: ", err)
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Cookie", BookIot.Token)
	resp, err := httpClient.Do(request)

	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		Logger.Error("Do BookPage request error: ", err)
	}

	if resp.StatusCode != http.StatusOK {
		errJson, _ := io.ReadAll(resp.Body)
		Logger.Errorf("BookPage response status error: %s, %s", resp.Status, string(errJson))
		return nil, fmt.Errorf("BookPage 请求出错，原因为: %s, %s", resp.Status, string(errJson))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resultCode := gjson.GetBytes(bodyBytes, "code").Int()
	if resultCode != 0 && resultCode != 200 {
		Logger.Errorf("BookPage response code error: %s",  string(bodyBytes))
		return nil, fmt.Errorf("获取数据错误，原因为: %s", bodyBytes)
	}

	pages := make([]BookPage, 0)
	dataResult := gjson.GetBytes(bodyBytes, "data.list")
	if !dataResult.Exists() {
		return pages, nil
	}
	err = json.Unmarshal([]byte(dataResult.String()), &pages)
	if err != nil {
		return nil, err
	}

	return pages, nil
}
