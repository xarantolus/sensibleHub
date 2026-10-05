package api

import (
	"errors"
	"log"
	"net/http"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
)

type ErrorCode string

const (
	CodeBadRequest        ErrorCode = "bad_request"
	CodeValidation        ErrorCode = "validation"
	CodeNotFound          ErrorCode = "not_found"
	CodeAlreadyDownloaded ErrorCode = "already_downloaded"
	CodeQueueFull         ErrorCode = "queue_full"
	CodeNotDownloading    ErrorCode = "not_downloading"
	CodeInternal          ErrorCode = "internal"
	CodeOther             ErrorCode = "other"
)

// Problem is the RFC 9457 body of every API error. Code is stable and meant
// for programmatic handling; Title and Detail are for humans.
type Problem struct {
	huma.ErrorModel
	Code     ErrorCode `json:"code" enum:"bad_request,validation,not_found,already_downloaded,queue_full,not_downloading,internal,other" doc:"Machine-readable error code"`
	Resource string    `json:"resource,omitempty" doc:"Kind of resource that was not found (song, album, artist)"`
	SongID   string    `json:"songId,omitempty" doc:"The existing song when the error is already_downloaded"`
}

func init() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		p := &Problem{Code: codeForStatus(status)}
		p.Status = status
		p.Title = http.StatusText(status)
		p.Detail = msg
		for _, err := range errs {
			if err != nil {
				p.Add(err)
			}
		}
		return p
	}
}

func codeForStatus(status int) ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnprocessableEntity:
		return CodeValidation
	case http.StatusNotFound:
		return CodeNotFound
	}
	if status >= 500 {
		return CodeInternal
	}
	return CodeOther
}

func newProblem(status int, code ErrorCode, detail string) *Problem {
	p := huma.NewError(status, detail).(*Problem)
	p.Code = code
	return p
}

// toProblem converts a store error into the API error a client can act on.
// Errors without a known meaning are logged and reported without internals.
func toProblem(op string, err error) error {
	var (
		notFound  *store.NotFoundError
		invalid   *store.ValidationError
		duplicate *store.AlreadyDownloadedError
		statusErr huma.StatusError
	)

	switch {
	case errors.As(err, &statusErr):
		return err
	case errors.As(err, &notFound):
		p := newProblem(http.StatusNotFound, CodeNotFound, notFound.Error())
		p.Resource = notFound.Kind
		return p
	case errors.As(err, &invalid):
		p := newProblem(http.StatusUnprocessableEntity, CodeValidation, invalid.Error())
		p.Add(&huma.ErrorDetail{Message: invalid.Reason, Location: "body." + invalid.Field})
		return p
	case errors.As(err, &duplicate):
		p := newProblem(http.StatusConflict, CodeAlreadyDownloaded, duplicate.Error())
		p.SongID = duplicate.SongID
		return p
	case errors.Is(err, store.ErrQueueFull):
		return newProblem(http.StatusTooManyRequests, CodeQueueFull, err.Error())
	case errors.Is(err, store.ErrNotDownloading):
		return newProblem(http.StatusConflict, CodeNotDownloading, err.Error())
	}

	log.Printf("[API] %s: %v\n", op, err)
	return newProblem(http.StatusInternalServerError, CodeInternal, "An unexpected error occurred, see the server log for details")
}
