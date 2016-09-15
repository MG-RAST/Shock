# creates statically compiled shock-server binary: /gopath/bin/shock-server

FROM golang:1.7.1-alpine

RUN apk update && apk add git make gcc libc-dev cyrus-sasl-dev

ENV GOPATH /go
COPY Makefile /go/
RUN cd /go && make install

CMD ["/bin/ash"]

