package api

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go-chess/rpc/user/client"
	"go-chess/rpc/user/model"
	"go-chess/util"
)

func register(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	question := ctx.PostForm("question")
	answer := ctx.PostForm("answer")
	uid := uuid.NewV4().String()

	if username == "" || password == "" {
		util.RespError(ctx, 400, "请将信息输入完整")
		return
	}

	l1 := len([]rune(username))
	l2 := len([]rune(password))
	l3 := len([]rune(question))
	l4 := len([]rune(answer))

	if l1 > 20 || l1 < 1 {
		util.RespErrorWithData(ctx, 400, "wrong length", "用户名请在1-8位之间")
		return
	}
	if l2 > 16 || l2 < 6 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密码请在6-16位之间")
		return
	}
	if l3 > 16 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密保问题请在16个字以内")
		return
	}
	if l4 > 10 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密保答案请在10个字以内")
		return
	}
	ctl := client.NewUserCtl("127.0.0.1:2379", "go-chess/user")
	res := ctl.CallRegister(model.User{
		Uuid:     uid,
		Name:     username,
		Password: password,
		Question: question,
		Answer:   answer,
	})
	if res.Status == true {
		util.RespSuccessful(ctx, res.Description)
		return
	}
	util.RespError(ctx, 400, res.Description)
	return
}

func login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if username == "" || password == "" {
		util.RespError(ctx, 400, "请将信息输入完整")
		return
	}

	l1 := len([]rune(username))
	l2 := len([]rune(password))

	if l1 > 20 || l1 < 1 {
		util.RespErrorWithData(ctx, 400, "wrong length", "用户名请在1-8位之间")
		return
	}
	if l2 > 16 || l2 < 6 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密码请在6-16位之间")
		return
	}

	ctl := client.NewUserCtl("127.0.0.1:2379", "go-chess/user")
	res := ctl.CallLogin(model.User{
		Name:     username,
		Password: password,
	})
	if res.Status == true {
		util.RespSuccessfulWithData(ctx, res.Description, res.Token)
		return
	}
	util.RespError(ctx, 400, res.Description)
	return
}

func changePassword(ctx *gin.Context) {
	username := ctx.PostForm("username")
	oldPassword := ctx.PostForm("old_password")
	newPassword := ctx.PostForm("new_password")

	if username == "" || oldPassword == "" || newPassword == "" {
		util.RespError(ctx, 400, "请将信息输入完整")
		return
	}

	l1 := len([]rune(username))
	l2 := len([]rune(oldPassword))
	l3 := len([]rune(newPassword))
	if l1 > 20 || l1 < 1 {
		util.RespErrorWithData(ctx, 400, "wrong length", "用户名请在1-8位之间")
		return
	}
	if l2 > 16 || l2 < 6 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密码请在6-16位之间")
		return
	}
	if l3 > 16 || l3 < 6 {
		util.RespErrorWithData(ctx, 400, "wrong length", "密码请在6-16位之间")
		return
	}

	ctl := client.NewUserCtl("127.0.0.1:2379", "go-chess/user")
	res := ctl.CallChangePW(model.ChangePassword{
		Name:        username,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	if res.Status == true {
		util.RespSuccessful(ctx, res.Description)
		return
	}
	util.RespError(ctx, 400, res.Description)
	return
}
