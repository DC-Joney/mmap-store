package statics

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"testing"
	"time"
)

func TestConfigRead(t *testing.T) {
	fmt.Printf("%#v", BookIot)
	path, err := url.JoinPath(BookIot.BaseURL, BookIot.BookInfo.Path)
	fmt.Println(url.JoinPath(BookIot.BaseURL, BookIot.BookInfo.Path))
	if err != nil {
		log.Fatalln("joinPath error： ", err)
	}

	fmt.Println(path)
}

func TestRequestStore(t *testing.T) {
	store, err := RequestStore("1", "1")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(store)

}

func TestRequestStoreTotal(t *testing.T) {
	total, _ := RequestStoreTotal()

	stores, err := RequestStore("1", fmt.Sprintf("%d", total))

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(stores)
}

func TestRequestStore2(t *testing.T) {

	total, err := RequestStoreTotal()

	if err != nil {
		log.Fatalln(err)
	}

	batchSize := 1 << 5

	count := int(total) / batchSize
	if int(total)&(batchSize-1) != 0 {
		count += 1
	}

	Logger.Infof("count: %d, batchSize: %d, total: %d", count, batchSize, total)

	request := NewBookStoreRequest(16)

	for i := 1; i < count; i++ {
		param := &BookStoreParam{
			pageNo:   i,
			pageSize: batchSize,
		}

		request.InvokeRequest(param)
	}

	ticker := time.NewTicker(2 * time.Second)

	for  {
		<-ticker.C
		Logger.Infof("Statics: %#v \n", Statics)
	}

}

func TestName(t *testing.T) {
	fmt.Println(math.Ceil(7 / 3))
}
