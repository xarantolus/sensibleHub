# All compilation and code generation goes through this file.
# Generated files (frontend/openapi.json, frontend/src/api/schema.ts) are committed;
# `make gen-check` fails when they are stale.

.PHONY: all gen gen-check frontend-deps frontend-build build server check check-go check-frontend clean

GO           ?= go
NPM          ?= npm
BINARY       ?= sensibleHub
GO_BUILD_FLAGS ?= -mod vendor -ldflags "-s -w"

GO_FILES     := $(shell find . -name '*.go' -not -path './vendor/*' -not -path './frontend/*' 2>/dev/null)
API_SOURCES  := $(shell find web/api store cmd/openapi -name '*.go' -not -name '*_test.go' 2>/dev/null)
FRONTEND_SRC := $(shell find frontend/src frontend/public -type f 2>/dev/null) \
                frontend/index.html frontend/vite.config.ts frontend/tsconfig.json \
                frontend/tsconfig.app.json frontend/tsconfig.node.json

OPENAPI_SPEC := frontend/openapi.json
API_TYPES    := frontend/src/api/schema.ts
OPENAPI_TS   := frontend/node_modules/.bin/openapi-typescript
OPENAPI_TS_FLAGS := --root-types --root-types-no-schema-prefix --root-types-keep-casing --immutable --enum-values

all: build

# --- code generation -------------------------------------------------------

$(OPENAPI_SPEC): $(API_SOURCES)
	$(GO) run -mod vendor ./cmd/openapi > $@.tmp
	mv $@.tmp $@

$(API_TYPES): $(OPENAPI_SPEC) frontend/node_modules/.installed
	$(OPENAPI_TS) $(OPENAPI_SPEC) -o $@ $(OPENAPI_TS_FLAGS)

gen: $(API_TYPES)

gen-check:
	$(GO) run -mod vendor ./cmd/openapi > $(OPENAPI_SPEC)
	$(OPENAPI_TS) $(OPENAPI_SPEC) -o $(API_TYPES) $(OPENAPI_TS_FLAGS)
	git diff --exit-code -- $(OPENAPI_SPEC) $(API_TYPES)

# --- frontend ---------------------------------------------------------------

frontend/node_modules/.installed: frontend/package.json frontend/package-lock.json
	cd frontend && $(NPM) ci
	touch $@

frontend-deps: frontend/node_modules/.installed

# Builds from the committed generated files; no Go toolchain needed (Docker frontend stage).
frontend-build: frontend/node_modules/.installed
	cd frontend && $(NPM) run build

frontend/dist/index.html: $(API_TYPES) $(FRONTEND_SRC) frontend/node_modules/.installed
	cd frontend && $(NPM) run build

# --- server -----------------------------------------------------------------

$(BINARY): frontend/dist/index.html $(GO_FILES) go.mod go.sum
	$(GO) build $(GO_BUILD_FLAGS) -o $@ .

build: $(BINARY)

# Compiles only the Go binary, embedding whatever frontend/dist holds (Docker, cross-compiles in pack.sh).
server:
	$(GO) build $(GO_BUILD_FLAGS) -o $(BINARY) .

# --- checks -----------------------------------------------------------------

check: check-go check-frontend

check-go:
	@test -z "$$(gofmt -l $(GO_FILES))" || { gofmt -l $(GO_FILES); echo "gofmt: files need formatting"; exit 1; }
	$(GO) vet -mod vendor ./...
	$(GO) test -mod vendor ./...

check-frontend: $(API_TYPES)
	cd frontend && $(NPM) run typecheck && $(NPM) run lint && $(NPM) test

clean:
	rm -rf frontend/dist $(BINARY) $(BINARY).exe releases $(OPENAPI_SPEC).tmp
