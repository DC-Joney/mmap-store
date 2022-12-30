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

// BookPagePiece 热区数据
type BookPagePiece struct {
	ApiKey string `json:"apiKey"`
	BookId int    `json:"bookId"`

	//热区对应的 pageNo
	PageNo int `json:"pageNo"`

	//热区索引
	PieceIndex int `json:"pieceIndex"`

	//是否定制
	Custom bool `json:"custom"`

	//热区文本
	PieceContent string `json:"pieceContent"`

	//热区多音频
	Audios []struct {
		Text    string `json:"text"`
		TextKey string `json:"textKey"`
	} `json:"multiplyAudioDTOList"`
}

type BookPagePieceParam struct {
	PageNo   int    `json:"pageNo"`
	BookSign string `json:"bookSign"`
	ApiKey   string `json:"apiKey"`
}

type BookPagePieceRequest struct {
	Pool     *ants.PoolWithFunc
	Receiver chan *BookPagePiece
}

//InvokeRequest 请求Store库
func (request *BookPagePieceRequest) InvokeRequest(param *BookPagePieceParam) {
	err := request.Pool.Invoke(param)
	if err != nil {
		Logger.Error("Add BookPagePieceParam to pools error: ", err)
	}
}

//requestStore 请求Store库
func (request *BookPagePieceRequest) workerRequest(data interface{}) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		Logger.Error("BookPiece worker pool error: ", err)
	//	}
	//}()

	param := data.(*BookPagePieceParam)
	bookPagePieces, err := RequestBookPagePiece(param)

	if err != nil {
		Logger.Errorf("workFunc==> 请求BookPiece 出错: param: %v, 错误原因：%s",param, err)
		return
	}

	for _, pagePiece := range bookPagePieces {
		Statics.AddPagePiece(&pagePiece)

		//todo 暂时不需要放入热区数据
		//request.Receiver <- &pagePiece
	}
}

func NewBookPagePieceRequest(poolSize int) *BookPagePieceRequest {
	request := &BookPagePieceRequest{
		Receiver: make(chan *BookPagePiece, 1<<9),
	}

	request.Pool, _ = ants.NewPoolWithFunc(poolSize, request.workerRequest)
	return request
}

func NewBookPagePieceParam(pageNo int, bookSign, apiKey string) *BookPagePieceParam {
	return &BookPagePieceParam{
		PageNo:   pageNo,
		BookSign: bookSign,
		ApiKey:   apiKey,
	}
}

//RequestBookPagePiece 请求绘本内页列表
// 	{"apiKey":"9f6fdcedf5c44111b5e14d632c043189","bookSign":"ac4f89488fbd443683cf0e9cd151652f","pageSize":6,"pageNo":1}
func RequestBookPagePiece(pageParam *BookPagePieceParam) ([]BookPagePiece, error) {

	pageUrl, err := url.JoinPath(BookIot.BaseURL, BookIot.BookPagePieceList.Path)
	if err != nil {
		Logger.Error("BookPiece join path error: ", err)
		return nil, err
	}

	method := strings.ToUpper(BookIot.BookPagePieceList.Method)
	requestJson, err := json.Marshal(pageParam)

	request, err := http.NewRequest(method, pageUrl, bytes.NewBuffer(requestJson))

	if err != nil {
		Logger.Error("BookPiece request error: ", err)
		return nil, err
	}

	//Logger.Infof("BookPieceRequest: method: %s, url: %s, body: %s", method, request.URL.String(), requestJson)

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Cookie", BookIot.Token)

	resp, err := httpClient.Do(request)
	defer func() {

		if resp == nil {
			Logger.Error("返回Response 错误，原因为: ", err)
		}

		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		Logger.Error("Do BookPiece request error: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errJson, _ := io.ReadAll(resp.Body)
		Logger.Errorf("BookPiece response status error: %s, %s", resp.Status, string(errJson))
		return nil, fmt.Errorf("BookPage 请求出错，原因为: %s, %s", resp.Status, string(errJson))
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	resultCode := gjson.GetBytes(bodyBytes, "code").Int()

	if resultCode != 0 && resultCode != 200 {
		Logger.Errorf("BookPiece response code error: %s",  string(bodyBytes))
		return nil, fmt.Errorf("获取数据错误，原因为: %s", bodyBytes)
	}

	pages := make([]BookPagePiece, 0)
	dataResult := gjson.GetBytes(bodyBytes, "data")
	if !dataResult.Exists() {
		return pages, nil
	}

	err = json.Unmarshal([]byte(dataResult.String()), &pages)
	if err != nil {
		return nil, err
	}

	return pages, nil
}
