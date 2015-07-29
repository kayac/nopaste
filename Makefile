GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: test get-deps binary install clean
test:
	go test

get-deps:
	go get -t -d -v .

binary:
	cd cmd/nopaste && gox -os="linux darwin" -arch="amd64" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE}"
	cd cmd/irc-msgr && gox -os="linux darwin" -arch="amd64" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE}"
	cd cmd/nopaste-cli && gox -os="linux darwin" -arch="amd64" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-X main.version ${GIT_VER} -X main.buildDate ${DATE} -X main.Endpoint $(ENDPOINT)"
	cd pkg && find . -name "*${GIT_VER}*" -type f -exec zip {}.zip {} \;

clean:
	rm -f pkg/*

all:
	cd cmd/nopaste && go build
	cd cmd/irc-msgr && go build

client:
	cd cmd/nopaste-cli && go build -ldflags "-X main.Endpoint $(ENDPOINT)"
