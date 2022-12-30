package statics

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"turing/resolve/tool"
)


type BookIotApiPath struct {

	//baseUrl
	BaseURL string `mapstructure:"baseUrl"`

	Token string `mapstructure:"cookie"`

	//查询绘本库的uri Path
	BookStore *HttpURL `mapstructure:"book-store"`

	//查询绘本库详情的 URI Path
	BookStoreDetail *HttpURL `mapstructure:"book-store-detail"`

	//绘本信息 URI Path
	BookInfo *HttpURL `mapstructure:"book-info"`

	//查看book 所有的页面
	BookPageList *HttpURL `mapstructure:"book-page-list"`

	//查看page 热区
	BookPagePieceList *HttpURL `mapstructure:"book-piece-list"`
}

type HttpURL struct {
	Path string `mapstructure:"path"`
	Method string `mapstructure:"method"`
}

const (
	pathResourceName = "book-iot.yaml"
)


var (
	BookIot *BookIotApiPath = new(BookIotApiPath)
)

func init() {
	resourceDir, err := tool.GetResourceDir()

	if err != nil {
		log.Fatalln("获取Resource Dir目录出错: ", err)
	}

	apiResourcePath := filepath.Join(resourceDir, pathResourceName)
	v := viper.New()

	v.SetConfigFile(apiResourcePath)

	if 	err := v.ReadInConfig(); err != nil{
		log.Fatalln("读取Yaml文件出错: ",err)
	}

	hookFunc := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		mapstructure.StringToSliceHookFunc(","),
	)

	bookIotViper := v.Sub("book-iot")
	err = bookIotViper.UnmarshalExact(BookIot, viper.DecodeHook(hookFunc))

	if err != nil {
		log.Fatalln("反序列化出错: ", err)
	}

}
