package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-chess/global"
	"go-chess/model"
	"log"
)

func InitConfig() {
	// 实例化viper
	v := viper.New()
	//文件的路径如何设置
	v.SetConfigFile("./setting-dev.yaml")
	err := v.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
	serverConfig := model.ServerConfig{}
	//给serverConfig初始值
	err = v.Unmarshal(&serverConfig)
	if err != nil {
		log.Println(err)
	}
	// 传递给全局变量
	global.Settings = serverConfig

	//热重载配置
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config file:%s Op:%s\n", e.Name, e.Op)
	})
	v.WatchConfig()
}
