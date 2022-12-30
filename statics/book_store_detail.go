package statics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// BookStoreDetail 绘本详情
type BookStoreDetail struct {

	// bookSign 值
	BookSign string `json:"bookSign"`
	Name     string `json:"name"`

	ApiKey string

	// bookSort: 6 卡片
	BookSort int64 `json:"bookSort"`

	//绘本内页数量
	InsidePageCount int `json:"insidePageCount"`

	//0 默认绘本，1 定制绘本
	Attribute int `json:"attribute"`

	// false 用户绘本， true 公有库绘本
	Custom bool `json:"custom"`
}

type BookStoreDetailParam struct {
	pageNo   int
	pageSize int
	storeId  string
	ctx      context.Context
}

type BookStoreDetailRequest struct {
	Pool         *ants.PoolWithFunc
	PageRequest  *BookPageRequest
	PieceRequest *BookPagePieceRequest
}

//InvokeRequest 请求Store库
func (request *BookStoreDetailRequest) InvokeRequest(param *BookStoreDetailParam) {
	err := request.Pool.Invoke(param)
	if err != nil {
		Logger.Error("Add store detail param to pools error: ", err)
	}
}

//requestStore 请求Store库
func (request *BookStoreDetailRequest) workerRequest(data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("StoreDetail worker pool error: ", err)
		}
	}()

	param := data.(*BookStoreDetailParam)
	pageNo := fmt.Sprintf("%d", param.pageNo)
	pageSize := fmt.Sprintf("%d", param.pageSize)
	apiKey := param.ctx.Value("apiKey").(string)

	bookStoreDetails, err := RequestStoreDetail(pageNo, pageSize, param.storeId)
	if err != nil {
		Logger.Info("workFunc==> 请求BookStoreDetail出错: ", err)
		return
	}

	for _, bookStoreDetail := range bookStoreDetails {
		bookStoreDetail.ApiKey = apiKey
		count := bookStoreDetail.InsidePageCount

		//Logger.Infof("绘本内页数量为: %d, custom: %t, attribute: %d", count, bookStoreDetail.Custom, bookStoreDetail.Attribute)

		//如果内页数量为0，或者是公有库的绘本则直接忽略
		if count == 0 || bookStoreDetail.Custom{
			continue
		}

		//针对非定制化的绘本才需要扫描内页图片以及内页多音频数据
		if bookStoreDetail.Attribute == 0  {
			bookPageParam := NewBookPageParam(count, bookStoreDetail.BookSign, apiKey)
			request.PageRequest.InvokeRequest(bookPageParam)
		}


		//请求绘本的热区数据，不管是否是定制绘本，内页热区都是需要搜索的
		for pageNo := 1; pageNo <= count; pageNo++ {
			pagePieceParam := NewBookPagePieceParam(pageNo, bookStoreDetail.BookSign, bookStoreDetail.ApiKey)
			//通过内页编号+bookSign + apiKey 获取热区数据
			request.PieceRequest.InvokeRequest(pagePieceParam)
		}
	}
}

func NewBookStoreDetailRequest(poolSize int) *BookStoreDetailRequest {
	request := &BookStoreDetailRequest{
		PageRequest:  NewBookPageRequest(poolSize),
		PieceRequest: NewBookPagePieceRequest(poolSize),
	}

	request.Pool, _ = ants.NewPoolWithFunc(poolSize, request.workerRequest)
	return request
}

func NewBookStoreDetailParam(count int, storeId string, ctx context.Context) *BookStoreDetailParam {
	return &BookStoreDetailParam{
		pageNo:   1,
		pageSize: count,
		storeId:  storeId,
		ctx:      ctx,
	}
}

//RequestStoreDetail 请求绘本列表库
func RequestStoreDetail(pageNo, pageSize, detailId string) ([]BookStoreDetail, error) {

	storeUrl, err := url.JoinPath(BookIot.BaseURL, BookIot.BookStoreDetail.Path)
	if err != nil {
		Logger.Error("StoreDetail join path error: ", err)
		return nil, err
	}

	method := strings.ToUpper(BookIot.BookStore.Method)
	request, err := http.NewRequest(method, storeUrl, nil)

	if err != nil {
		Logger.Error("StoreDetail store request error: ", err)
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Cookie", BookIot.Token)

	query := request.URL.Query()

	query.Add("bookStoreDetailId", detailId)
	query.Add("pageNo", pageNo)
	query.Add("pageSize", pageSize)

	request.URL.RawQuery = query.Encode()
	resp, err := httpClient.Do(request)
	defer func() {
		_ = resp.Body.Close()
	}()

	if err != nil {
		Logger.Error("Do StoreDetail request error: ", err)
	}

	if resp.StatusCode != http.StatusOK {
		errJson, _ := io.ReadAll(resp.Body)
		Logger.Errorf("StoreDetail response status error: %s, %s", resp.Status, string(errJson))
		return nil, fmt.Errorf("请求出错，原因为: %s, %s", resp.Status, string(errJson))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resultCode := gjson.GetBytes(bodyBytes, "code").Int()

	if resultCode != 0 {
		Logger.Errorf("Store response code error: %s",  string(bodyBytes))
		return nil, fmt.Errorf("获取数据错误，原因为: %s", bodyBytes)
	}

	details := make([]BookStoreDetail, 0)
	dataResult := gjson.GetBytes(bodyBytes, "data.list")

	if !dataResult.Exists() {
		return details, nil
	}

	err = json.Unmarshal([]byte(dataResult.String()), &details)

	if err != nil {
		return nil, err
	}

	return details, nil
}
