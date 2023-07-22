FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/mirai -ldflags '-s -w' .

FROM alpine:latest

RUN set -ex && \
  apk update && \
  apk upgrade && \
  rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin/mirai /usr/local/bin/mirai

CMD [ "mirai" ]
