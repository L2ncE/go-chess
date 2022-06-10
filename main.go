package main

import (
	"fmt"
	"go-chess/config"
	mysql "go-chess/dao/gorm"
	"go-chess/dao/redis"
	"log"
)

func main() {
	config.InitConfig()

	if err := mysql.InitGormDB(); err != nil {
		log.Printf("init gorm failed, err:%v\n", err)
	} else {
		log.Println("连接GORM成功!")
	}

	if err := redis.InitRedis(); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
	} else {
		log.Println("连接Redis成功!")
	}
}
