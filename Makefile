.PHONY: all tidy build install windows clean run docker

VERSION = $(shell git describe --tags --always --dirty)
BUILD = $(shell date +%FT%T%z)

OUT = mirai
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.build=$(BUILD)
GOFLAGS = -ldflags '$(LDFLAGS)'
export CGO_ENABLED = 1

all: tidy build
tidy:
	go mod tidy
	go mod vendor
build:
	go build -o "$(OUT)" $(GOFLAGS)
install:
	go install $(GOFLAGS)
windows:
	GOOS=windows \
	GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc \
	go build -o "$(OUT).exe" $(GOFLAGS)
clean:
	go clean
	rm -f "$(OUT)"
	rm -f "$(OUT).exe"
test:
	go test
run:
	go run .
docker:
	sudo docker build --pull -t cloudwindy/mirai:latest .
