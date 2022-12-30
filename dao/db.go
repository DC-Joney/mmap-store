package dao

import (
	"gorm.io/gorm"
	"turing/resolve/config"
)

var (
	DB *gorm.DB
)

func init()  {

	//初始化配置
	config.InitConfig()




}

