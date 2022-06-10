package mysql

import (
	"go-chess/rpc/user/model"
	"log"
)

func InsertUser(user model.User) error {
	deres := db.Select("uuid", "name", "password", "question", "answer").
		Create(&model.User{Uuid: user.Uuid, Name: user.Name, Password: user.Password, Question: user.Question, Answer: user.Answer})
	err := deres.Error
	if err != nil {
		log.Printf("insert failed, err:%v\n", err)
		return err
	}
	return err
}

func SelectUserByName(name string) error {
	var user model.User
	dbRes := db.Model(&model.User{}).Where("name = ?", name).First(&user)
	err := dbRes.Error
	if err != nil {
		log.Println("select failed, err:", err)
		return err
	}
	return nil
}

func SelectPasswordByName(name string) (string, error) {
	var user model.User
	dbRes := db.Model(&model.User{}).Select("password").Where("name = ?", name).First(&user)
	err := dbRes.Error
	if err != nil {
		log.Println("select failed, err:", err)
		return "", err
	}
	return user.Password, nil
}

func SelectUuIdByName(name string) (string, error) {
	user := model.User{}
	err := db.Model(&model.User{}).Select("uuid").Where("name = ?", name).Find(&user).Error
	if err != nil {
		log.Println("select failed, err:", err)
		return "", err
	}
	return user.Uuid, nil
}
