package mysql

import (
	"go-chess/model"
	"log"
)

func SelectUserNameByUUId(uuid string) (string, error) {
	var user model.User
	dbRes := db.Model(&model.User{}).Select("name").Where("uuid = ?", uuid).First(&user)
	err := dbRes.Error
	if err != nil {
		log.Println("select failed, err:", err)
		return "", err
	}
	return user.Name, nil
}
