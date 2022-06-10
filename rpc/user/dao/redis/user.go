package redis

import (
	"log"
)

func IsRegister(name string) (bool, error) { //当前用户名是否被注册
	flag, err := rdb.SIsMember("username", name).Result()
	if err != nil {
		log.Println("redis judge is register err:", err)
		return false, err
	}
	return flag, nil
}

func AddUsername(name string) error {
	err := rdb.SAdd("username", name).Err()
	if err != nil {
		log.Println("redis add name err", err)
		return err
	}
	return nil
}

func IsUsernameCacheAlive() (bool, error) {
	es, err := rdb.SMembers("username").Result()
	if err != nil {
		log.Println("openid cache get error:", err)
		return false, err
	}
	if len(es) > 0 {
		return true, nil
	}
	return false, nil
}

func Ping() (bool, error) {
	_, err := rdb.Ping().Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
