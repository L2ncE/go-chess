# go-chess
红岩网校工作站2022春季后端期末考核：双人象棋

## 目录

- [go-chess](#go-chess)
    - [目录](#目录)
    - [功能实现](功能实现)
        - [基础功能](基础功能)   
        - [加分项](加分项)
            - [功能类](功能类)
            - [技术类](技术类)
    - [接口说明](#接口说明)
    - [加分项实现](#加分项实现)
    - [快速开始](#快速开始)

##  功能实现

### 基础功能

- 用户登陆注册更改密码

- 加入房间
  - 玩家能主动开房，能够指定房间号加入房间

- 同一房间最多两人进入
- 房间内
  - 玩家可以切换准备和未准备状态

- 两玩家都处于准备状态时可以自行开启游戏（本地）
- 游戏对战
  - 以坐标的形式处理棋子位置信息

- 固定的棋盘大小

- 双方轮流着棋

- 判断获胜条件，在获胜条件出现后结束游戏

- 部署

  - 使用**Docker**部署

- 客户端显示效果

  - 使用Ebiten第三方库实现

### 加分项

#### 功能类

- 房间
  - 房间内玩家聊天
  - 对低俗玩家踢出房间（说脏话超过三次）
  - 可以多房间同时进行，一名用户也可以同时进入多个房间

#### 技术类

- 服务拆分，用户中心使用**gRPC**重构，并使用**etcd**进行服务发现
- 使用**redis**缓存，使服务能够承受更高的负载
- 撰写**Dockerfile**生成仅有18m的镜像
- 使用**pprof**进行性能调优
- 使用**Viper**进行项目配置，并支持热重载配置
- 使用**cron**定时任务进行无用缓存的删除

## 接口说明

### 注册 POST `42.192.155.29:6666/user/register`

BODY

| KEY      | DESCRIPTION |
| -------- | ----------- |
| username | 必填        |
| password | 必填        |
| question | 可选        |
| answer   | 可选        |

### 登录 POST  `42.192.155.29:6666/user/login`

BODY

| KEY      | DESCRIPTION |
| -------- | ----------- |
| username | 必填        |
| password | 必填        |

### 改密码 PUT `42.192.155.29:6666/user/password`

HEADER

| KEY   | DESCRIPTION |
| ----- | ----------- |
| TOKEN | 必填        |

BODY

| KEY          | DESCRIPTION |
| ------------ | ----------- |
| username     | 必填        |
| old_password | 必填        |
| new_password | 必填        |

### 切换准备状态 GET `42.192.155.29:6666/ready/:room_id`

HEADER

| KEY   | DESCRIPTION |
| ----- | ----------- |
| TOKEN | 必填        |

PARAM

| KEY     | DESCRIPTION |
| ------- | ----------- |
| room_id | 必填        |

### 加入房间 WebSocket `ws://42.192.155.29:6666/?room_id=red`

HEADER

| KEY   | DESCRIPTION |
| ----- | ----------- |
| TOKEN | 必填        |

PARAM

| KEY     | DESCRIPTION |
| ------- | ----------- |
| room_id | 必填        |

## 加分项实现

### gRPC + ETCD

撰写`proto`文件

```protobuf
syntax = "proto3";

package user;

option go_package = "./user";

message RegisterReq{
  string username = 1;
  string password = 2;
  string question = 3;
  string answer = 4;
  string uuid = 5;
}

message RegisterRes{
  bool status = 1;
  string description = 2;
}

message LoginReq{
  string username = 1;
  string password = 2;
}

message LoginRes{
  bool status = 1;
  string token = 2;
  string description = 3;
}

message changeReq {
  string old_password = 1;
  string new_password = 2;
  string username = 3;
}

message changeRes {
  bool status = 1;
  string description = 2;
}

service UserCenter{
  rpc Register(RegisterReq) returns (RegisterRes);
  rpc Login(LoginReq) returns(LoginRes);
  rpc changePW (changeReq) returns (changeRes);
}
```

通过./一个cmd文件自动生成代码

```shell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./user.proto
```

接下来就是进行分层，然后分别撰写server.go以及client.go并server中将服务加入到etcd中

### redis缓存 cron定时任务

在房间模块中使用redis，当你选择进入房间后会在redis中产生此房间的集合（并且最多能容纳两个人），退出时会将你从房间中删掉，定时任务会每个小时扫一遍，房间为空则删除。下次又有用户进入此房间的话则再次生成此房间集合

准备状态也是使用redis实现，也就是使用的最经典的点赞系统那套，点一次是准备，再点一次就取消，是根据集合中是否有你这个用户而实现的（依赖集合中成员不可重复的特性）

**样例**

```go
//cron
_, err := c.AddFunc("@every 1h", func() {
   err := redis.DeleteEmptyRoom()
   if err != nil {
      log.Println("cron err", err)
      return
   }
})
```

```go
//redis
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
```

### Docker

就是摈弃掉多余的内容，使用镜像

```dockerfile
FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/main ./main.go


FROM scratch

ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/main /app/main

EXPOSE 6666
CMD ["./main"]
```

### pprof性能优化

```go
func InitPprofMonitor() {
   go func() {
      log.Println(http.ListenAndServe(":9990", nil))
   }()
}
```

进入界面 http://localhost:9990/debug/pprof/

![image-20220612113027339](https://s2.loli.net/2022/06/12/3z7eBc6XuWr8IPq.png)

此节目可以查询详细参数，若想更加可视化一点可以安装graphviz，并在命令行中输入

```shell
$ go tool pprof -http=:8080 "http://localhost:9990/debug/pprof/heap //或其他
```

![image-20220612113241221](https://s2.loli.net/2022/06/12/Bh1XPjErqcLVHzI.png)

可以更加直观，或者进入火焰图界面，效果也很好

![image-20220612113314188](https://s2.loli.net/2022/06/12/ASmjbZ3ovwqKVGd.png)



参考 https://link.juejin.cn/?target=https%3A%2F%2Fgithub.com%2Fwolfogre%2Fgo-pprof-practice

### Viper配置

使用viper配置并进行热重载设置

配置模板 `setting-dev.yaml`

```yaml
# settings-dev.yaml
name: "go-chess"
port: 1234

gorm:
  name: "xx"
  host: "xx.xx.xx.xx"
  port: 3306
  password: "xxxxxxxx"
  dbName: "xx"

redis:
  host: "xx.xx.xx.xx"
  port: 6379
  password: ":.xxxx@:xxx?Zx"
  DB: 10

etcd:
  addr: "xxxx.0.x:2379"
```

```go
func InitConfig() {
   // 实例化viper
   v := viper.New()
   //文件的路径如何设置
   v.SetConfigFile("./setting-dev.yaml")
   err := v.ReadInConfig()
   if err != nil {
      log.Println(err)
   }
   serverConfig := model.ServerConfig{}
   //给serverConfig初始值
   err = v.Unmarshal(&serverConfig)
   if err != nil {
      log.Println(err)
   }
   // 传递给全局变量
   global.Settings = serverConfig

   //热重载配置
   v.OnConfigChange(func(e fsnotify.Event) {
      log.Printf("config file:%s Op:%s\n", e.Name, e.Op)
   })
   v.WatchConfig()
```

## 快速开始

