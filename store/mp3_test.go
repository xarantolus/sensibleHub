package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/music"
)

func editableSong(id string, lastEdit time.Time) music.Entry {
	e := song(id, true, lastEdit)
	e.LastEdit = lastEdit
	e.MusicData.Title = id
	e.MusicData.Duration = 100
	e.AudioSettings.End = 100
	return e
}

func retitle(t *testing.T, m *Manager, id string) {
	t.Helper()
	if _, err := m.UpdateEntry(id, SongEdit{Title: "New title", End: 100, Sync: true}); err != nil {
		t.Fatal(err)
	}
}

func knownMP3Size(m *Manager, id string) (int64, bool) {
	e := m.Songs[id]
	return e.MP3Size()
}

func writeGeneratedMP3(t *testing.T, id string, usedAt time.Time) string {
	t.Helper()
	p := music.MP3FilePath(id)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, usedAt, usedAt); err != nil {
		t.Fatal(err)
	}
	return p
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func TestMP3SizeIsKnownUntilTheSongChanges(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(editableSong("a", time.Now()))

	if _, ok := knownMP3Size(m, "a"); ok {
		t.Fatal("a song that was never generated has no known size")
	}

	m.recordMP3Size("a", m.Songs["a"].LastEdit, 1234)
	if size, ok := knownMP3Size(m, "a"); !ok || size != 1234 {
		t.Fatalf("size = %d, %v; want 1234", size, ok)
	}

	retitle(t, m, "a")
	if _, ok := knownMP3Size(m, "a"); ok {
		t.Error("an edit must invalidate the remembered size")
	}

	e := m.Songs["a"]
	e.MP3 = &music.GeneratedMP3{Size: 1234, LastEdit: e.LastEdit, Version: music.MP3Version - 1}
	if _, ok := e.MP3Size(); ok {
		t.Error("a size from an older generator version must not be used")
	}
}

func TestMP3SizeOfAnOutdatedGenerationIsIgnored(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(editableSong("a", time.Now()))

	m.recordMP3Size("a", m.Songs["a"].LastEdit.Add(-time.Minute), 1234)
	if _, ok := knownMP3Size(m, "a"); ok {
		t.Error("a file generated before the last edit must not set the size")
	}
}

func TestMP3SizeIsSaved(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(editableSong("a", time.Now()))
	m.recordMP3Size("a", m.Songs["a"].LastEdit, 1234)
	m.SaveMP3Sizes()

	loaded, err := NewManager(m.cfg)
	if err != nil {
		t.Fatal(err)
	}
	if size, ok := knownMP3Size(loaded, "a"); !ok || size != 1234 {
		t.Errorf("size after reload = %d, %v; want 1234", size, ok)
	}
}

func TestEditRemovesGeneratedMP3(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(editableSong("a", time.Now()))
	p := writeGeneratedMP3(t, "a", time.Now())

	retitle(t, m, "a")
	if exists(p) {
		t.Error("the MP3 from before the edit is still there")
	}
}

func TestDeleteGeneratedFilesByLastUse(t *testing.T) {
	t.Chdir(t.TempDir())
	// Both songs were edited long ago; only when their MP3 was last used matters.
	old := time.Now().Add(-30 * 24 * time.Hour)
	m := testManager(editableSong("stale", old), editableSong("recent", old))
	stale := writeGeneratedMP3(t, "stale", time.Now().Add(-4*24*time.Hour))
	recent := writeGeneratedMP3(t, "recent", time.Now().Add(-time.Hour))

	if n := m.DeleteGeneratedFiles(-1); n != 0 || !exists(stale) {
		t.Fatal("a negative age must keep every file")
	}

	if n := m.DeleteGeneratedFiles(3); n != 1 {
		t.Errorf("deleted %d files, want 1", n)
	}
	if exists(stale) || !exists(recent) {
		t.Errorf("stale kept: %v, recent kept: %v; want only the recent one kept", exists(stale), exists(recent))
	}

	if n := m.DeleteGeneratedFiles(0); n != 1 || exists(recent) {
		t.Error("an age of 0 must delete every file")
	}
}
