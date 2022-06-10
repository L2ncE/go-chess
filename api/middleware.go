package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go-chess/rpc/user/model"
	"go-chess/util"
)

var mySigningKey = []byte("go-chess")

// JWTAuth JWT登录
func JWTAuth(ctx *gin.Context) {
	token := ctx.Request.Header.Get("token")
	if token == "" {
		util.RespErrorWithData(ctx, 400, "jwt get info error", "Without your information, please log in first!")
		ctx.Abort()
		return
	}
	ctx.Set("uuid", ParseToken(token))
	ctx.Next()
}

func ParseToken(s string) string {
	//解析传过来的token
	tokenClaims, err := jwt.ParseWithClaims(s, &model.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return tokenClaims.Claims.(*model.MyClaims).Uuid
}

//CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, token, x-access-token")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}
