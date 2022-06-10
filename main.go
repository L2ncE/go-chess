package main

import (
	"go-chess/api"
	"go-chess/config"
)

func main() {
	config.InitConfig()
	api.InitEngine()
}
