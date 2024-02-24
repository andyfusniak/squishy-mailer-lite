OUTPUT_DIR=$(shell go env GOPATH)/bin
GIT_COMMIT=$(shell git rev-parse --short HEAD)
VERSION=v0.1.0

all: sqm

sqm:
	@GCO_ENABLED=1 go build -o $(OUTPUT_DIR)/sqm -ldflags "-X 'main.version=${VERSION}' -X 'main.gitCommit=${GIT_COMMIT}'" ./cmd/sqm/main.go

.PHONY: clean
clean:
	-@rm -r $(OUTPUT_DIR)/sqm 2> /dev/null || true
