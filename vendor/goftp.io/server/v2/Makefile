DIST := dist
GO ?= go
SHASUM ?= shasum -a 256

export PATH := $($(GO) env GOPATH)/bin:$(PATH)

GOFILES := $(shell find . -name "*.go" -type f ! -path "*/bindata.go")
SERVER_VERSION ?= $(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')
SERVER_VERSION_TAG ?= $(shell sed 's/+/_/' <<< $(SERVER_VERSION))

TAGS ?=
LDFLAGS := -X "gitea.com/goftp/server.version=$(SERVER_VERSION)" -s -w

# override to allow passing additional goflags via make CLI
override GOFLAGS := $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)'

PACKAGES ?= $(shell $(GO) list ./...)
SOURCES ?= $(shell find . -name "*.go" -type f)

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: install-lint-tools
	$(GO) run github.com/mgechev/revive@v1.3.2 -config .revive.toml ./... || exit 1

.PHONY: misspell-check
misspell-check: install-lint-tools
	$(GO) run github.com/client9/misspell/cmd/misspell@latest -error -i unknwon,destory $(GOFILES)

.PHONY: misspell
misspell: install-lint-tools
	$(GO) run github.com/client9/misspell/cmd/misspell@latest -w -i unknwon $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$(go fmt ./...); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: test
test:
	$(GO) test $(PACKAGES)

.PHONY: unit-test-coverage
unit-test-coverage:
	$(GO) test -cover -coverprofile coverage.out $(PACKAGES) && echo "\n==>\033[32m Ok\033[m\n" || exit 1

.PHONY: tidy
tidy:
	$(GO) mod tidy

install-lint-tools:
	@hash revive > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/mgechev/revive@v1.3.2; \
	fi
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/client9/misspell/cmd/misspell@latest; \
	fi
