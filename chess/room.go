package main

import (
	"log"
)

func ReadyNum(roomId string) (error, int) {
	es, err := rdb.SMembers("ready_" + roomId).Result()
	if err != nil {
		log.Println("ready num cache get error:", err)
		return err, -1
	}
	return nil, len(es)
}
