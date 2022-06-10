package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

var rdb *redis.Client

func InitRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", "42.192.155.29", 16379), // redis地址
		Password: ":.Bc:W2D@:cb~xrRf)?Z",                       // redis密码，没有则留空
		DB:       10,                                           // 默认数据库，默认是0
	})

	//通过 *redis.Client.Ping() 来检查是否成功连接到了redis服务器
	_, err = rdb.Ping().Result()
	if err != nil {
		log.Printf("连接失败：%v\n", err)
		return err
	}
	return nil
}
