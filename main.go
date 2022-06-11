package main

import "go-chess/chess"

func main() {
	//config.InitConfig()
	//
	//if err := mysql.InitGormDB(); err != nil {
	//	log.Printf("init gorm failed, err:%v\n", err)
	//} else {
	//	log.Println("连接GORM成功!")
	//}
	//
	//if err := redis.InitRedis(); err != nil {
	//	fmt.Printf("init redis failed, err:%v\n", err)
	//} else {
	//	log.Println("连接Redis成功!")
	//}
	//
	//api.InitEngine()
	chess.NewGame()
}
