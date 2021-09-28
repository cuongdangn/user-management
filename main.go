package main

import (
	"fmt"
	"user-management/config"
	"user-management/coreserver"
	"user-management/log"
	"user-management/webserver"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.NewLog()
	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	csv, err := coreserver.NewServer(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	go csv.Start()

	wsv, err := webserver.NewWebServer(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = wsv.Start(cfg.WebServer.Host + ":" + cfg.WebServer.Port)
	if err != nil {
		fmt.Println(err)
	}
}
