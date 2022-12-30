package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"turing/resolve/tool"
)

type DataSource struct {
	DriverName string `mapstructure:"driverName"`
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Charset    string `mapstructure:"charset"`
	Loc        string `mapstructure:"loc"`
}

const (
	profileEnv        = "profile.active"
	applicationConfig = "application-%s.yml"
	defaultEnv        = "local"
)

var (
	profile    = defaultEnv
	datasource *DataSource
)

func InitConfig() {
	resourceDir, err := tool.GetResourceDir()
	if err != nil {
		log.Fatalln("Get resource dir error: ", err)
	}

	env, b := os.LookupEnv(profileEnv)
	if b {
		profile = env
	}

	configName := fmt.Sprintf(applicationConfig, profile)
	configPath := filepath.Join(resourceDir, configName)

	log.Println("configPath: ", configPath)

	v := viper.NewWithOptions()
	v.SetConfigFile(configPath)

	hookFunc := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		mapstructure.StringToSliceHookFunc(","),
	)

	if err := v.ReadInConfig(); err != nil {
		log.Fatal("Read config error: ", err)
	}

	v.SetDefault("application.name", "turing")

	dataSourceSub := v.Sub("datasource")

	datasource = &DataSource{}
	err = dataSourceSub.Unmarshal(datasource, viper.DecodeHook(hookFunc))

	if err != nil {
		log.Fatalln("Yaml unmarshal error: ", err)
	}
}
