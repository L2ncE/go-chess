package main

import (
	"go-chess/api"
	"go-chess/config"
	"go-chess/dao/mysql"
	"go-chess/dao/redis"
	"go-chess/pprof"
	"go-chess/task"
	"log"
)

func main() {
	config.InitConfig()
	pprof.InitPprofMonitor()

	if err := mysql.InitGormDB(); err != nil {
		log.Printf("init gorm failed, err:%v\n", err)
	} else {
		log.Println("连接GORM成功!")
	}

	if err := redis.InitRedis(); err != nil {
		log.Printf("init redis failed, err:%v\n", err)
	} else {
		log.Println("连接Redis成功!")
	}

	task.CronInit()

	api.InitEngine()
}
