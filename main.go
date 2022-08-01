package main

import (
	"embed"
	"flag"
	"log"
	"os/exec"

	"xarantolus/sensibleHub/ftp"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/web"

	// Supported image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

var (
	//go:embed templates
	templateFS embed.FS

	//go:embed assets
	assetFS embed.FS
)

func main() {
	var (
		flagDebug  = flag.Bool("debug", false, "Start the server in debug mode")
		flagConfig = flag.String("config", config.DefaultConfigFile, "Path to the configuration file")
	)
	flag.Parse()

	if *flagDebug {
		log.Println("[Debug] Debug mode enabled")
	}

	// Let's load our config file
	cfg, err := config.Parse(*flagConfig)
	if err != nil {
		panic("while parsing config: " + err.Error())
	}

	// Check all external commands that are used by this program in order to warn the user if they aren't accessible
	checkInstalledCommand := func(cmd string) {
		_, err := exec.LookPath(cmd)
		if err != nil {
			log.Printf("[Warning] Cannot find %s (%s). Please make sure it is installed and that the path is correct.\n", cmd, err.Error())
		}
	}
	checkInstalledCommand(cfg.Alternatives.FFmpeg)
	checkInstalledCommand(cfg.Alternatives.FFprobe)
	checkInstalledCommand(cfg.Alternatives.YoutubeDL)

	// Let's initialize our Manager. This is the main data structure
	// that handles basically everything
	manager, err := store.NewManager(cfg)
	if err != nil {
		panic("while initializing manager: " + err.Error())
	}

	// At first, we check if there are any songs in `import/` that
	// we could move to our music collection
	err = manager.ImportFiles("import")
	if err != nil {
		log.Printf("Error while importing: %s\n", err.Error())
	}

	// Clean up all unused data on disk.
	// Also if a song directory was deleted we delete it from our dataset
	n := manager.CleanUp()
	if n == 0 {
		log.Println("[Cleanup] No cleanup necessary")
	} else {
		if n == 1 {
			log.Println("[Cleanup] Finished cleaning up", n, "song")
		} else {
			log.Println("[Cleanup] Finished cleaning up", n, "songs")
		}
	}

	// Kick off cover preview generation.
	// That way they aren't all generated on the first load of the /songs page
	go manager.GenerateCoverPreviews()

	// Start the FTP server
	go func() {
		err := ftp.RunServer(manager, cfg)
		if err != nil {
			panic("while running ftp server: " + err.Error())
		}
	}()

	// And the web server of course
	err = web.RunServer(manager, cfg, assetFS, templateFS, *flagDebug)
	if err != nil {
		panic("while running web server: " + err.Error())
	}
}
