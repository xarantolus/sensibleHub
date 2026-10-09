package store

import (
	"fmt"
	"log"
)

// migrations[i] upgrades the data from schema version i to i+1. A migration only
// reshapes stored fields; expensive derived data is filled in by background jobs
// such as StartAnalysis.
var migrations = []func(m *Manager) error{
	// 0 → 1: entries may now carry an optional analysis; nothing to convert.
	func(*Manager) error { return nil },
	// 1 → 2: optional unsynced_artists list. Nothing to convert, but an older
	// build would drop the list when saving, so it must refuse this file.
	func(*Manager) error { return nil },
}

func currentSchema() int { return len(migrations) }

// migrate brings data loaded from an older version up to date. Before the first
// change it copies the data file to <file>.bak-v<old version>.
func (m *Manager) migrate() error {
	from := m.SchemaVersion
	if from == currentSchema() {
		return nil
	}
	if from > currentSchema() {
		return fmt.Errorf("data file has schema version %d, but this build only knows up to %d; refusing to run an older sensibleHub on newer data", from, currentSchema())
	}

	backup := fmt.Sprintf("%s.bak-v%d", managerDataFile, from)
	if err := copyOverwrite(managerDataFile, backup); err != nil {
		return fmt.Errorf("backing up data before migration: %w", err)
	}

	for v := from; v < currentSchema(); v++ {
		if err := migrations[v](m); err != nil {
			return fmt.Errorf("migrating data from schema %d to %d: %w", v, v+1, err)
		}
		m.SchemaVersion = v + 1
	}
	log.Printf("[Startup] Migrated data from schema %d to %d (backup: %s)\n", from, m.SchemaVersion, backup)
	return m.Save()
}
