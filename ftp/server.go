package ftp

import (
	"log"

	"xarantolus/sensiblehub/store/config"

	"github.com/goftp/server"
)

// RunServer runs the FTP server until it crashes
func RunServer(cfg config.Config) (err error) {
	opts := &server.ServerOpts{
		Factory: &musicDriverFactory{},
		Port:    cfg.FTP.Port,
		Auth:    &server.SimpleAuth{Name: cfg.FTP.User, Password: cfg.FTP.Passwd},
	}

	server := server.NewServer(opts)

	log.Printf("[FTP] Server listening on port %d\n", cfg.FTP.Port)
	return server.ListenAndServe()
}
