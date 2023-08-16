FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
# also, prepare task for building
COPY go.mod go.sum ./
RUN set -ex && \
  go mod download && \
  go mod verify && \
  go install github.com/go-task/task/v3/cmd/task@latest

COPY . .
RUN set -ex && \
  OUT=/usr/local/bin/mirai task build

FROM alpine:latest

RUN set -ex && \
  apk update && \
  apk upgrade && \
  rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin/mirai /usr/local/bin/mirai

CMD [ "mirai" ]
