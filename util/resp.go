package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RespSuccessful(ctx *gin.Context, description interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":        200,
		"description": description,
	})
}

func RespSuccessfulWithData(ctx *gin.Context, description interface{}, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":        200,
		"description": description,
		"data":        data,
	})
}

func RespError(ctx *gin.Context, code int, description interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":        code,
		"description": description,
	})
}

func RespErrorWithData(ctx *gin.Context, code int, description interface{}, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":        code,
		"description": description,
		"data":        data,
	})
}
