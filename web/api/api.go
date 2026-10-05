// Package api is the typed HTTP API of sensibleHub. Its OpenAPI document is the
// source the frontend's client is generated from (see cmd/openapi).
package api

import (
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humamux"
	"github.com/gorilla/mux"
)

const base = "/api/v1"

type body[T any] struct {
	Body T
}

// New registers all API operations on r. m may be nil when the API is only
// built to generate its OpenAPI document.
func New(r *mux.Router, m *store.Manager) huma.API {
	huma.DefaultArrayNullable = false

	cfg := huma.DefaultConfig("sensibleHub", "1.0.0")
	cfg.OpenAPIPath = base + "/openapi"
	cfg.DocsPath = base + "/docs"
	cfg.SchemasPath = base + "/schemas"
	cfg.CreateHooks = nil

	api := humamux.New(r, cfg)

	registerSongs(api, m)
	registerLibrary(api, m)
	registerDownloads(api, m)
	registerEvents(api, m)

	return api
}
