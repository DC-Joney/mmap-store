package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
)

const (
	bookPageTableName = "book_page_multiply_audio"
)

var (
	userDataApi = ""
	bookDetailPath = "/book_info"
)

type BookPage struct {
	BookSign  string
	PageNo    string
	AudioText string
}

type BookPageResponse struct {
	BookSign  string
	PageNo    string
	AudioText string
}

// FindPage 查询bookSign相关的内页多音频
func (b *BookPage) FindPageMultiAudio(bookSign string) []BookPage{
	bookPage := DB.Session(&gorm.Session{}).Table(bookPageTableName)
	pages := make([]BookPage, 0)
	bookPage.Select("book_sign,page_no,audio_text").Where("book_sign=?", bookSign).Scan(pages)
	return pages
}

// FindPage 查询bookSign相关的内页信息
func (b *BookPage) FindPage(bookSign,apiKey string) ([]BookPage, error){

	searchParam := struct {
		BookSign string `json:"book_sign"`
		ApiKey string `json:"api_key"`
	}{}

	marshal, err := json.Marshal(searchParam)

	if err != nil {
		log.Println("")
	}

	reader := bytes.NewReader(marshal)

	request, err := http.NewRequest(http.MethodPost, userDataApi, reader)
	request.Header.Add("Content-Type","application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	return nil, nil
}
