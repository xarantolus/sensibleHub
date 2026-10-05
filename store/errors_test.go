package store

import (
	"errors"
	"fmt"
	"testing"
	"xarantolus/sensibleHub/store/music"
)

func TestTypedErrorsMatchSentinels(t *testing.T) {
	tests := []struct {
		name string
		err  error
		is   error
		not  error
	}{
		{"not found", songNotFound("abcd"), ErrNotFound, ErrInvalid},
		{"validation", &ValidationError{Field: "title", Reason: "empty"}, ErrInvalid, ErrNotFound},
		{"already downloaded", &AlreadyDownloadedError{SongID: "abcd", SongName: "A - B"}, ErrAlreadyKnown, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := fmt.Errorf("context: %w", tt.err)
			if !errors.Is(wrapped, tt.is) {
				t.Errorf("wrapped error should match %v", tt.is)
			}
			if errors.Is(wrapped, tt.not) {
				t.Errorf("wrapped error should not match %v", tt.not)
			}
		})
	}
}

func TestNotFoundErrorCarriesWhatWasMissing(t *testing.T) {
	var nf *NotFoundError
	if !errors.As(fmt.Errorf("x: %w", songNotFound("abcd")), &nf) {
		t.Fatal("errors.As should find NotFoundError")
	}
	if nf.Kind != "song" || nf.Key != "abcd" {
		t.Errorf("got kind %q key %q", nf.Kind, nf.Key)
	}
}

func TestDuplicateDownloadErrorExposesExistingSong(t *testing.T) {
	err := error(&DownloadError{
		URL:    "https://example.com/x",
		Reason: DownloadDuplicate,
		Err:    &AlreadyDownloadedError{SongID: "abcd", SongName: "A - B"},
	})

	if !errors.Is(err, ErrAlreadyKnown) {
		t.Error("duplicate download should match ErrAlreadyKnown")
	}
	var dup *AlreadyDownloadedError
	if !errors.As(err, &dup) || dup.SongID != "abcd" {
		t.Errorf("errors.As should reach the duplicate's song, got %v", dup)
	}

	other := &DownloadError{URL: "u", Reason: DownloadNoAudio, Err: errors.New("no audio")}
	if errors.Is(other, ErrAlreadyKnown) {
		t.Error("a non-duplicate failure must not match ErrAlreadyKnown")
	}
}

func TestSubscriberReceivesEventsInOrder(t *testing.T) {
	m := &Manager{}
	events, cancel := m.Subscribe()
	defer cancel()

	m.publish(SongAdded{Song: music.Entry{ID: "a"}})
	m.publish(SongDeleted{ID: "a"})

	if got, ok := (<-events).(SongAdded); !ok || got.Song.ID != "a" {
		t.Errorf("first event should be SongAdded for a, got %v", got)
	}
	if got, ok := (<-events).(SongDeleted); !ok || got.ID != "a" {
		t.Errorf("second event should be SongDeleted for a, got %v", got)
	}
}

func TestSlowSubscriberIsDroppedInsteadOfBlocking(t *testing.T) {
	m := &Manager{}
	slow, cancelSlow := m.Subscribe()
	defer cancelSlow()
	fast, cancelFast := m.Subscribe()
	defer cancelFast()

	const published = 200
	received := 0
	for i := 0; i < published; i++ {
		m.publish(SongDeleted{ID: "x"})
		for len(fast) > 0 {
			<-fast
			received++
		}
	}

	if received != published {
		t.Errorf("a reading subscriber should get all %d events, got %d", published, received)
	}

	buffered := 0
	for range slow {
		buffered++
	}
	if buffered == 0 || buffered >= published {
		t.Errorf("slow subscriber should get its buffered events and then a closed channel, got %d", buffered)
	}
}

func TestCancelAfterOverflowDoesNotPanic(t *testing.T) {
	m := &Manager{}
	_, cancel := m.Subscribe()
	for i := 0; i < 100; i++ {
		m.publish(SongDeleted{ID: "x"})
	}
	cancel()
	cancel()
}
