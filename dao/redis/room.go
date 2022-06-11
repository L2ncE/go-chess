package redis

import (
	"log"
)

func AddRoom(id string, uuid string) error {
	err := rdb.SAdd("room_"+id, uuid).Err()
	if err != nil {
		log.Println("redis add uuid err", err)
		return err
	}
	return nil
}

func AddRoomId(id string) error {
	err := rdb.SAdd("room", id).Err()
	if err != nil {
		log.Println("redis add room err", err)
		return err
	}
	return nil
}

func IsAliveRoom(id string) (bool, error) { //当前用户是否注册
	flag, err := rdb.SIsMember("room", id).Result()
	if err != nil {
		log.Println("redis judge is alive err:", err)
		return false, err
	}
	return flag, nil
}

func RoomNum(id string) (int, error) {
	es, err := rdb.SMembers("room_" + id).Result()
	if err != nil {
		log.Println("openid cache get error:", err)
		return -1, err
	}
	return len(es), nil
}

func ReadySet(roomId string, uuid string) (error, int) {
	val, err := rdb.SIsMember("ready_"+roomId, uuid).Result()
	if err != nil {
		log.Println("judge ready error:", err)
		return err, 2
	}
	if val == false {
		_, err := rdb.SAdd("ready_"+roomId, uuid).Result()
		if err != nil {
			log.Println("set ready error:", err)
			return err, 2
		}
		return nil, 1
	} else {
		_, err := rdb.SRem("ready_"+roomId, uuid).Result()
		if err != nil {
			log.Println("set cancel ready error:", err)
			return err, 2
		}
		return nil, 0
	}
}

func IsInRoom(roomId string, uuid string) (bool, error) { //当前用户是否在此房间
	flag, err := rdb.SIsMember("room_"+roomId, uuid).Result()
	if err != nil {
		log.Println("redis judge is in the room err:", err)
		return false, err
	}
	return flag, nil
}

func DeleteEmptyRoom() error {
	set, err := rdb.SMembers("room").Result()
	if err != nil {
		log.Println("len of room get error:", err)
		return err
	}
	for _, v := range set {
		es, err := rdb.SMembers("room_" + v).Result()
		if err != nil {
			log.Println(err)
			return err
		}
		if len(es) <= 0 {
			rdb.Del("room_" + v)
		}
	}
	return nil
}
