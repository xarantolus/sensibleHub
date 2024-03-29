package ftp

import (
	"crypto/subtle"
	"fmt"
	"log"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"

	"goftp.io/server/core"
)

// RunServer runs the FTP server until it crashes
func RunServer(manager *store.Manager, cfg config.Config) (err error) {
	opts := &core.ServerOpts{
		Factory: &musicDriverFactory{
			Manager: manager,
		},
		Port:   cfg.FTP.Port,
		Auth:   &configAuth{cfg: cfg},
		Logger: &core.DiscardLogger{},
	}

	server := core.NewServer(opts)

	log.Printf("[FTP] Server listening on port %d\n", cfg.FTP.Port)
	return server.ListenAndServe()
}

// configAuth implements the server.Auth interface
type configAuth struct {
	cfg config.Config
}

// CheckPasswd will check user's password
func (a *configAuth) CheckPasswd(name, pass string) (bool, error) {
	for _, user := range a.cfg.FTP.Users {
		if constantTimeEquals(name, user.Name) && constantTimeEquals(pass, user.Passwd) {
			log.Printf("[FTP] Logged in %s\n", user.Name)
			return true, nil
		}
	}
	return false, fmt.Errorf("login failure")
}

func constantTimeEquals(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
