package ftp

import (
	"crypto/subtle"
	"fmt"
	"log"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"

	"goftp.io/server/v2"
)

// RunServer runs the FTP server until it crashes
func RunServer(manager *store.Manager, cfg config.Config) (err error) {
	s, err := server.NewServer(&server.Options{
		Driver: &musicDriver{manager: manager},
		Port:   cfg.FTP.Port,
		Auth:   &configAuth{cfg: cfg},
		Perm:   server.NewSimplePerm("human", "music"),
		Logger: &server.DiscardLogger{},
	})
	if err != nil {
		return err
	}
	s.RegisterNotifer(&treeSnapshot{manager: manager})

	log.Printf("[FTP] Server listening on port %d\n", cfg.FTP.Port)
	return s.ListenAndServe()
}

// configAuth implements the server.Auth interface
type configAuth struct {
	cfg config.Config
}

// CheckPasswd will check user's password
func (a *configAuth) CheckPasswd(_ *server.Context, name, pass string) (bool, error) {
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
