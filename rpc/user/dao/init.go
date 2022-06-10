package dao

import (
	"fmt"
	mysql "go-chess/rpc/user/dao/gorm"
	"go-chess/rpc/user/dao/redis"
	"log"
)

func Init() {
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
