package main

import (
	"go-chess/api"
	"go-chess/config"
	"go-chess/pprof"
)

func main() {
	config.InitConfig()
	pprof.InitPprofMonitor()
	api.InitEngine()
}
