FROM gcc:latest

RUN set -ex && \
  apt-get update && \
  apt-get -y install gccgo-go && \
  apt-get clean

COPY build modules *.go go.* /usr/src/mirai/
WORKDIR /usr/src/mirai

RUN set -ex && \
  go build -compiler gccgo .