package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-chess/global"
	"log"
)

func InitEngine() {
	engine := gin.Default()
	engine.Use(CORS())
	go h.run()

	userGroup := engine.Group("/user")
	{
		userGroup.POST("/register", register)
		userGroup.POST("/login", login)
		{
			userGroup.Use(JWTAuth)
			userGroup.PUT("/password", changePassword)
		}
	}

	wsGroup := engine.Group("/")
	{
		wsGroup.Use(JWTAuth)
		wsGroup.GET("/", serverWs)
		wsGroup.GET("/ready/:room_id", ready)
	}

	err := engine.Run(fmt.Sprintf(":%d", global.Settings.Port))
	if err != nil {
		log.Printf("init error:%v\n", err)
		return
	}
}
