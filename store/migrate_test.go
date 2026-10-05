package store

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"xarantolus/sensibleHub/store/config"
)

const legacyData = `{"songs": {"abcd": {"id": "abcd"}}}`

func writeDataFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managerDataFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNewManagerMigratesLegacyDataFile(t *testing.T) {
	t.Chdir(t.TempDir())
	writeDataFile(t, legacyData)

	m, err := NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.GetEntry("abcd"); !ok {
		t.Error("song from the legacy file was not loaded")
	}
	if m.SchemaVersion != currentSchema() {
		t.Errorf("schema version %d, want %d", m.SchemaVersion, currentSchema())
	}

	if got := string(readFile(t, managerDataFile+".bak-v0")); got != legacyData {
		t.Errorf("backup differs from the original: %q", got)
	}

	var saved struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(readFile(t, managerDataFile), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.SchemaVersion == nil || *saved.SchemaVersion != currentSchema() {
		t.Errorf("saved schema_version = %v, want %d", saved.SchemaVersion, currentSchema())
	}
}

func TestNewManagerDoesNotMigrateTwice(t *testing.T) {
	t.Chdir(t.TempDir())
	writeDataFile(t, legacyData)
	if _, err := NewManager(config.Config{}); err != nil {
		t.Fatal(err)
	}

	before, err := os.Stat(managerDataFile)
	if err != nil {
		t.Fatal(err)
	}
	content := readFile(t, managerDataFile)

	m, err := NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.GetEntry("abcd"); !ok {
		t.Error("song missing after reload")
	}

	after, err := os.Stat(managerDataFile)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Error("data file was rewritten although it was current")
	}
	if string(readFile(t, managerDataFile)) != string(content) {
		t.Error("data file content changed")
	}
	if _, err := os.Stat(managerDataFile + ".bak-v1"); err == nil {
		t.Error("a second backup was created")
	}
}

func TestNewManagerRefusesNewerSchema(t *testing.T) {
	t.Chdir(t.TempDir())
	newer := `{"schema_version": 999, "songs": {}}`
	writeDataFile(t, newer)

	_, err := NewManager(config.Config{})
	if err == nil {
		t.Fatal("expected an error for data written by a newer version")
	}
	if !strings.Contains(err.Error(), "999") {
		t.Errorf("error should name the data's schema version: %v", err)
	}

	if got := string(readFile(t, managerDataFile)); got != newer {
		t.Errorf("data file was modified: %q", got)
	}
	entries, err := os.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("unexpected files next to the data file: %v", entries)
	}
}

func TestNewManagerFreshInstallNeedsNoMigration(t *testing.T) {
	t.Chdir(t.TempDir())

	m, err := NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != currentSchema() {
		t.Errorf("schema version %d, want %d", m.SchemaVersion, currentSchema())
	}
	if _, err := os.Stat("data"); err == nil {
		t.Error("a fresh install must not create files before the first save")
	}

	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	m2, err := NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if m2.SchemaVersion != currentSchema() {
		t.Errorf("reloaded schema version %d", m2.SchemaVersion)
	}
	if _, err := os.Stat(managerDataFile + ".bak-v0"); err == nil {
		t.Error("a current data file must not be backed up")
	}
}
