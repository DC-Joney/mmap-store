package statics

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
)

type BookStore struct {
	StoreId   string `json:"id"`
	ApiKey    string `json:"storeNumber"`
	BookCount int    `json:"bookCount,string"`
}

func (b *BookStore) UnmarshalJSON(bytes []byte) error {
	result := gjson.ParseBytes(bytes)
	b.StoreId = result.Get("id").String()
	b.ApiKey = result.Get("storeNumber").String()
	count := result.Get("bookCount").Int()
	b.BookCount = int(count)
	return nil
}

type BookStoreParam struct {
	pageNo   int
	pageSize int
}

type BookStoreRequest struct {
	poolFunc      *ants.PoolWithFunc
	DetailRequest *BookStoreDetailRequest
}

//InvokeRequest 请求Store库
func (request *BookStoreRequest) InvokeRequest(param *BookStoreParam) {
	err := request.poolFunc.Invoke(param)
	if err != nil {
		Logger.Error("Add store param to pools error: ", err)
	}
}

func (request *BookStoreRequest) workerFunc(data interface{}) {

	defer func() {
		if err:= recover(); err != nil {
			Logger.Error("Store worker pool error: ", err)
		}
	}()


	param := data.(*BookStoreParam)
	pageNo := fmt.Sprintf("%d", param.pageNo)
	pageSize := fmt.Sprintf("%d", param.pageSize)
	bookStores, err := RequestStore(pageNo, pageSize)

	if err != nil {
		Logger.Info("workFunc==> 请求BookStore出错: ", err)
		return
	}

	for _, bookStore := range bookStores {
		//如果没有绘本则不需要发起请求
		count := bookStore.BookCount
		if count == 0 {
			continue
		}

		ctx := context.WithValue(context.TODO(), "apiKey", bookStore.ApiKey)
		detailParam := NewBookStoreDetailParam(count, bookStore.StoreId, ctx)
		request.DetailRequest.InvokeRequest(detailParam)
	}
}

func (request BookStoreRequest) PieceReceiver() <-chan *BookPagePiece {
	return request.DetailRequest.PieceRequest.Receiver
}

func (request BookStoreRequest) PageReceiver() <-chan *BookPage {
	return request.DetailRequest.PageRequest.Receiver
}

func NewBookStoreRequest(poolSize int) *BookStoreRequest {
	request := &BookStoreRequest{
		DetailRequest: NewBookStoreDetailRequest(poolSize),
	}

	request.poolFunc, _ = ants.NewPoolWithFunc(poolSize, request.workerFunc)
	return request
}

func NewStoreParam(count int) *BookStoreParam {
	return &BookStoreParam{
		pageNo:   1,
		pageSize: count,
	}
}

func RequestStore(pageNo, pageSize string) ([]BookStore, error) {
	storeJson, err := requestStoreJson(pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(storeJson)
	stores := make([]BookStore, 0)
	result.Get("data.list").ForEach(func(_, value gjson.Result) bool {
		store := BookStore{
			StoreId:   value.Get("id").String(),
			ApiKey:    value.Get("storeNumber").String(),
			BookCount: int(value.Get("bookCount").Int()),
		}
		stores = append(stores, store)
		return true
	})

	return stores, nil
}

//RequestStore 请求绘本列表库
func requestStoreJson(pageNo, pageSize string) ([]byte, error) {

	storeUrl, err := url.JoinPath(BookIot.BaseURL, BookIot.BookStore.Path)
	if err != nil {
		Logger.Error("Store join path error: ", err)
		return nil, err
	}

	request, err := http.NewRequest(BookIot.BookStore.Method, storeUrl, nil)

	if err != nil {
		Logger.Error("Create store request error: ", err)
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Cookie", BookIot.Token)

	query := request.URL.Query()

	query.Add("keyword", "")
	query.Add("pageNo", pageNo)
	query.Add("pageSize", pageSize)

	request.URL.RawQuery = query.Encode()

	Logger.Infof("Store Request url: %s", request.URL.String())

	resp, err := httpClient.Do(request)
	defer func() {
		_ = resp.Body.Close()
	}()

	if err != nil {
		Logger.Error("Do store request error: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errJson, _ := io.ReadAll(resp.Body)
		Logger.Errorf("Store response status error: %s, %s", resp.Status, string(errJson))
		return nil, fmt.Errorf("请求出错，原因为: %s, %s", resp.Status, string(errJson))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resultCode := gjson.GetBytes(bodyBytes, "code")

	if !resultCode.Exists() || resultCode.Int() != 0 {
		Logger.Errorf("Store response code error: %s",  string(bodyBytes))
		return nil, fmt.Errorf("获取数据错误，原因为: %s", bodyBytes)
	}

	return bodyBytes, nil
}

// RequestStoreTotal 请求总共的total数量
func RequestStoreTotal() (int64, error) {
	store, err := requestStoreJson("1", "1")
	if err == nil {
		Logger.Infof("Get Store Request json: %s", store)
		totalValue := gjson.GetBytes(store, "data.total")
		if totalValue.Exists() {
			return totalValue.Int(), nil
		}
	}

	return 0, err
}
