OS   = "linux"
ARCH = "amd64"

BINARY_NAME := homyaksrv
MAIN_PACKAGE_PATH := ./cmd/server/main.go
MODULE_NAME := github.com/victorzhuk/homyak

APP_VERSION     := $(or $(APP_VERSION),local)
APP_COMMIT_LONG ?= -
APP_BUILD_TIME  ?= -

APP_LDFLAGS="-s -w \
	-X $(MODULE_NAME)/internal/app.Version=$(APP_VERSION) \
	-X $(MODULE_NAME)/internal/app.Commit=$(APP_COMMIT_LONG) \
	-X $(MODULE_NAME)/internal/app.BuildAt=$(APP_BUILD_TIME)"

.PHONY: deps
tidy:
	go mod tidy -v
	go fmt ./...

.PHONY: test
test:
	go test ./... -v -race -timeout 30s -buildvcs

.PHONY: bench
bench:
	go test ./... -v -race -bench . -benchtime 5s -timeout 0 -run=XXX -cpu 1 -benchmem

.PHONY: lint
lint:
	go mod verify
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --config=.golangci.yml ./...

.PHONY: lint
lint-changes:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --new-from-rev=HEAD~1 --config=.golangci.yml ./...

.PHONY: build
build:
	go mod download
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags $(APP_LDFLAGS) -o=./bin/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

.PHONY: run
run:
	go run $(MAIN_PACKAGE_PATH)

.PHONY: gen
gen:
	go generate ./...
