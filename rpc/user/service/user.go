package service

import (
	mysql "go-chess/rpc/user/dao/gorm"
	"go-chess/rpc/user/dao/redis"
	"go-chess/rpc/user/model"
	"gorm.io/gorm"
)

func IsRepeatUsername(name string) (bool, error) {
	flag, err := redis.IsUsernameCacheAlive()
	if err != nil {
		return true, err
	}
	if flag { //有缓存
		flag, err = redis.IsRegister(name) //查看是否注册过
		if err != nil {
			return true, err
		}
		return flag, nil
	}
	err = mysql.SelectUserByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound { //找不到会报这个错误捏
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func InsertUser(user model.User) error {
	err := mysql.InsertUser(user)
	return err
}

func IsPasswordCorrect(username, password string) (bool, error) {
	searchPassword, err := mysql.SelectPasswordByName(username)
	if err != nil {
		return false, err
	}
	if password == searchPassword {
		return true, nil
	}
	return false, nil
}

func GetUuidByUsername(username string) (string, error) {
	uuid, err := mysql.SelectUuIdByName(username)
	return uuid, err
}
