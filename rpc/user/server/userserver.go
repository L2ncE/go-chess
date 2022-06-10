package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go-chess/etcd"
	"go-chess/rpc/user/dao"
	"go-chess/rpc/user/model"
	"go-chess/rpc/user/pd"
	"go-chess/rpc/user/service"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type server struct {
	user.UnimplementedRegisterCenterServer
}

var etcdCenter = "127.0.0.1:2379"

func etcdInit(config *etcd.Config) (err error) {
	nodeServer, err := etcd.NewNodeServer(config)
	if err != nil {
		return err
	}

	err = nodeServer.StartServer()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	dao.Init()
	addr := flag.String("addr", "err", "127.0.0.1:50001")
	flag.Parse()
	listen, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Println(err)
	}
	s := grpc.NewServer()
	user.RegisterRegisterCenterServer(s, &server{})

	err = etcdInit(&etcd.Config{
		EleName:   "go-chess/user",      //选举的信息
		NodeName:  *addr,                //节点的val
		Endpoints: []string{etcdCenter}, //etcd的服务地址
	})
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("正在监听......")
	if err = s.Serve(listen); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) Register(_ context.Context, req *user.RegisterReq) (res *user.RegisterRes, err error) {
	userInfo := model.User{
		Name:     req.Username,
		Password: req.Password,
		Question: req.Question,
		Answer:   req.Answer,
		Uuid:     req.Uuid,
	}
	res = &user.RegisterRes{}

	isRepeat, err := service.IsRepeatUsername(req.Username)
	if err != nil {
		log.Println(err)
		res.Status = false
		res.Description = "judge is repeat username failed:" + err.Error()
		return res, err
	}

	if isRepeat {
		res.Status = false
		res.Description = "name is repeated"
		return res, nil
	}

	err = service.InsertUser(userInfo)
	if err != nil {
		log.Println(err)
		res.Status = false
		res.Description = "register user failed:" + err.Error()
		return res, err
	}
	res.Status = true
	res.Description = "register successful"
	return res, nil
}

func (s *server) Login(_ context.Context, req *user.LoginReq) (res *user.LoginRes, err error) {
	isC, err := service.IsPasswordCorrect(req.Username, req.Password)
	if err != nil {
		log.Println(err)
		return &user.LoginRes{
			Status:      false,
			Description: "judge password failed",
		}, err
	}

	if isC {
		uuid, err := service.GetUuidByUsername(req.Username)

		c := model.MyClaims{
			Uuid: uuid,
			StandardClaims: jwt.StandardClaims{
				NotBefore: time.Now().Unix() - 60,
				ExpiresAt: time.Now().Unix() + 2592000, //30天，仅做测试
				Issuer:    "YuanXinHao",
			},
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
		var mySigningKey = []byte("go-chess")
		s, err := t.SignedString(mySigningKey)
		if err != nil {
			log.Println(err)
			return &user.LoginRes{
				Status:      false,
				Description: "get token failed",
			}, err
		}
		return &user.LoginRes{
			Status:      true,
			Description: "login successful",
			Token:       s,
		}, nil
	}
	return &user.LoginRes{
		Status:      false,
		Description: "wrong password",
	}, nil
}
