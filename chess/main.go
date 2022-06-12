package main

import (
	"fmt"
	"log"
)

func main() {
	var roomId string
	InitConfig()
	if err := InitRedis(); err != nil {
		log.Printf("init redis failed, err:%v\n", err)
	} else {
		log.Println("init redis success!")
	}
	for {
		fmt.Printf("Enter exit to exit,Please input your roomId:")
		_, err := fmt.Scanf("%s", &roomId)
		if err != nil {
			log.Println("scan error, err:", err)
			return
		}
		if roomId == "exit" {
			break
		}
		err, readyNum := ReadyNum(roomId)
		if readyNum == 2 {
			NewGame()
		} else {
			fmt.Println("the room is not full")
		}
	}
	fmt.Println("game over")
}
