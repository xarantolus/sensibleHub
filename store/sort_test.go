package store

import (
	"fmt"
	"testing"
	"time"

	"xarantolus/sensibleHub/store/music"
)

func songsAt(n int, start time.Time) []music.Entry {
	var out []music.Entry
	for i := 0; i < n; i++ {
		out = append(out, song(fmt.Sprintf("s%d", i), true, start.Add(-time.Duration(i)*time.Minute)))
	}
	return out
}

func TestNewestReturnsAtLeast25(t *testing.T) {
	m := testManager(songsAt(40, time.Now().Add(-72*time.Hour))...)
	list, today := m.Newest()
	if len(list) != 25 || today {
		t.Fatalf("got %d songs, today=%v; want 25, false", len(list), today)
	}
	if list[0].ID != "s0" {
		t.Fatalf("first = %s, want s0", list[0].ID)
	}
}

func TestNewestFewSongs(t *testing.T) {
	m := testManager(songsAt(3, time.Now().Add(-72*time.Hour))...)
	list, today := m.Newest()
	if len(list) != 3 || today {
		t.Fatalf("got %d songs, today=%v; want 3, false", len(list), today)
	}
}

func TestNewestReturnsAllOfTodayWhenMoreThan25(t *testing.T) {
	m := testManager(songsAt(30, time.Now().Add(24*time.Hour))...)
	list, today := m.Newest()
	if len(list) != 30 || !today {
		t.Fatalf("got %d songs, today=%v; want 30, true", len(list), today)
	}
}
