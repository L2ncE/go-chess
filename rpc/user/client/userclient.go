package client

import (
	"context"
	"fmt"
	"go-chess/etcd"
	"go-chess/rpc/user/model"
	user "go-chess/rpc/user/pd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type userInfo struct {
	uuid        string
	username    string
	password    string
	question    string
	answer      string
	oldPassword string
	newPassword string
	node        string
	conn        *grpc.ClientConn
}

func NewUserCtl(endpoint, serverName string) (u *userInfo) {
	kv, err := etcd.NewClient(etcd.ConfigEtcdAddr{EtcdAddr: endpoint}).MatchAServer(serverName)
	fmt.Println("使用节点", kv)
	if err != nil {
		log.Println(err)
		return &userInfo{}
	}

	conn, err := grpc.Dial(kv.Val, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
	}
	userCtl := userInfo{
		node: kv.Val,
		conn: conn,
	}
	return &userCtl
}

func (u *userInfo) CallRegister(info model.User) *user.RegisterRes {
	u.uuid = info.Uuid
	u.username = info.Name
	u.password = info.Password
	u.question = info.Question
	u.answer = info.Answer
	client := user.NewUserCenterClient(u.conn)
	return u.register(client)
}

func (u *userInfo) CallLogin(info model.User) *user.LoginRes {
	u.username = info.Name
	u.password = info.Password
	client := user.NewUserCenterClient(u.conn)
	return u.login(client)
}

func (u *userInfo) CallChangePW(info model.ChangePassword) *user.ChangeRes {
	u.username = info.Name
	u.oldPassword = info.OldPassword
	u.newPassword = info.NewPassword
	client := user.NewUserCenterClient(u.conn)
	return u.changePW(client)
}

func (u *userInfo) register(client user.UserCenterClient) *user.RegisterRes {
	res, err := client.Register(context.Background(), &user.RegisterReq{
		Uuid:     u.uuid,
		Username: u.username,
		Password: u.password,
		Question: u.question,
		Answer:   u.answer,
	})
	if err != nil {
		log.Println(err)
		return &user.RegisterRes{}
	}
	if res.Status == true {
		return res
	}
	return &user.RegisterRes{}
}

func (u *userInfo) login(client user.UserCenterClient) *user.LoginRes {
	res, err := client.Login(context.Background(), &user.LoginReq{
		Username: u.username,
		Password: u.password,
	})
	if err != nil {
		log.Println(err)
		return &user.LoginRes{
			Status: false,
			Token:  "",
		}
	}
	return res
}

func (u *userInfo) changePW(client user.UserCenterClient) *user.ChangeRes {
	res, err := client.ChangePW(context.Background(), &user.ChangeReq{
		Username:    u.username,
		OldPassword: u.oldPassword,
		NewPassword: u.newPassword,
	})
	if err != nil {
		log.Println(err)
		return &user.ChangeRes{
			Status:      false,
			Description: "error:" + err.Error(),
		}
	}
	return res
}
