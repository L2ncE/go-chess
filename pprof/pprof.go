package pprof

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

// 可视化 http://localhost:9990/debug/pprof/
//go tool pprof -http=:5557 "http://localhost:9990/debug/pprof/heap"
//go tool pprof -http=:8080 "http://localhost:9990/debug/pprof/heap"

func InitPprofMonitor() {
	go func() {
		log.Println(http.ListenAndServe(":9990", nil))
	}()
}
