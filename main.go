package main

import (
	"log"
	"xarantolus/sensiblehub/ftp"
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

	err = store.M.ImportFiles("import")
	if err != nil {
		log.Printf("Error while importing: %s\n", err.Error())
	}

	n := store.M.CleanUp()
	if n == 0 {
		log.Println("No cleanup necessary")
	} else {
		if n == 1 {
			log.Println("Finished cleaning up", n, "song")
		} else {
			log.Println("Finished cleaning up", n, "songs")
		}
	}

	go func() {
		err := ftp.RunServer(cfg)
		if err != nil {
			panic("while running ftp server: " + err.Error())
		}
	}()

	err = web.RunServer(cfg)
	if err != nil {
		panic("while running web server: " + err.Error())
	}
}
