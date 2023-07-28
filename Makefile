.PHONY: all build clean run docker

OUT = mirai
LDFLAGS = -s -w
GOFLAGS = -ldflags '$(LDFLAGS)'
export CGO_ENABLED=1

all: build
build:
	go build -o "$(OUT)" $(GOFLAGS)
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
	docker build --pull -t cloudwindy/mirai:latest .