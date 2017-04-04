# creates statically compiled shock-server binary: /go/bin/shock-server

FROM golang:1.7.5-alpine

RUN apk update && apk add git make gcc libc-dev cyrus-sasl-dev

ENV GOPATH /go
COPY Makefile /go/
RUN cd /go && make install

CMD ["/bin/ash"]

