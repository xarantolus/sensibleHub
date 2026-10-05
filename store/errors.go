package store

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalid        = errors.New("invalid input")
	ErrAlreadyKnown   = errors.New("already downloaded")
	ErrQueueFull      = errors.New("cannot enqueue more songs at this time")
	ErrNotDownloading = errors.New("no download is running")
)

type NotFoundError struct {
	Kind string
	Key  string
}

func (e *NotFoundError) Error() string        { return fmt.Sprintf("%s %q not found", e.Kind, e.Key) }
func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }

func songNotFound(id string) error { return &NotFoundError{Kind: "song", Key: id} }

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string        { return e.Field + ": " + e.Reason }
func (e *ValidationError) Is(target error) bool { return target == ErrInvalid }

type AlreadyDownloadedError struct {
	SongID   string
	SongName string
}

func (e *AlreadyDownloadedError) Error() string {
	return fmt.Sprintf("%q has already been downloaded", e.SongName)
}
func (e *AlreadyDownloadedError) Is(target error) bool { return target == ErrAlreadyKnown }

type DownloadFailure string

const (
	DownloadAborted      DownloadFailure = "aborted"
	DownloadToolFailed   DownloadFailure = "tool_failed"
	DownloadNoAudio      DownloadFailure = "no_audio"
	DownloadInvalidAudio DownloadFailure = "invalid_audio"
	DownloadDuplicate    DownloadFailure = "duplicate"
	DownloadInternal     DownloadFailure = "internal"
)

// DownloadError describes why a queued download did not produce a song.
// Output holds the downloader's combined output when it is relevant for diagnosis.
type DownloadError struct {
	URL    string
	Reason DownloadFailure
	Output string
	Err    error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("download of %s failed (%s): %v", e.URL, e.Reason, e.Err)
}
func (e *DownloadError) Unwrap() error { return e.Err }
