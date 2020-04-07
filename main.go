package main

import (
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/config"
	"xarantolus/sensiblehub/web"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		panic("while parsing config: " + err.Error())
	}

	err = store.InitializeManager()
	if err != nil {
		panic("while initializing manager: " + err.Error())
	}

	err = web.RunServer(cfg)
	if err != nil {
		panic("while running web server: " + err.Error())
	}
}
