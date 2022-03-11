GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
export GO111MODULE := on

.PHONY: test build clean client
test:
	go test

build:
	goreleaser build --snapshot --rm-dist

clean:
	rm -rf dist/*

client:
	cd cmd/nopaste-cli && go build -ldflags "-X main.Endpoint=$(ENDPOINT)" .
