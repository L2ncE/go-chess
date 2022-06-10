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

	err := engine.Run(fmt.Sprintf(":%d", global.Settings.Port))
	if err != nil {
		log.Printf("init error:%v\n", err)
		return
	}
}