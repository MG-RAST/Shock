
# creates statically compiled shock-server binary: /gopath/bin/shock-server

FROM golang:1.6.3-alpine

RUN apk update && apk add git make gcc

ENV GOROOT /usr/local/go 
ENV PATH /usr/local/go/bin:/gopath/bin:/usr/local/bin:$PATH 
ENV GOPATH /go/

RUN mkdir -p /gopath/ && \
  cd /gopath/ && \
  curl -s -O https://raw.githubusercontent.com/MG-RAST/Shock/master/Makefile && \
  make install
