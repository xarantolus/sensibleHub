package main

import (
	"flag"
	"log"
	"os/exec"
	"xarantolus/sensibleHub/ftp"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/web"
)

var (
	flagDebug = flag.Bool("debug", false, "Start the server in debug mode")
)

func main() {
	flag.Parse()

	if *flagDebug {
		log.Println("[Debug] Debug mode enabled")
	}

	var checkInstalledCommand = func(cmd string) {
		_, err := exec.LookPath(cmd)
		if err != nil {
			log.Printf("[Warning] Cannot find %s (%s). Please make sure to install it.\n", cmd, err.Error())
		}
	}
	// Check all external commands that are used by this program in order to warn the user if they aren't accessible
	checkInstalledCommand("ffmpeg")
	checkInstalledCommand("youtube-dl")

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

	n := store.M.CleanUp(cfg)
	if n == 0 {
		log.Println("[Cleanup] No cleanup necessary")
	} else {
		if n == 1 {
			log.Println("[Cleanup] Finished cleaning up", n, "song")
		} else {
			log.Println("[Cleanup] Finished cleaning up", n, "songs")
		}
	}

	go func() {
		err := ftp.RunServer(cfg)
		if err != nil {
			panic("while running ftp server: " + err.Error())
		}
	}()

	err = web.RunServer(cfg, *flagDebug)
	if err != nil {
		panic("while running web server: " + err.Error())
	}
}
