
# docker build -t mgrast/shock -f Dockerfile_new .
# docker run --rm --name test -ti mgrast/shock /bin/ash

FROM golang:1.7.1-alpine

RUN apk update && apk add gcc libc-dev cyrus-sasl-dev

COPY . /go/src/github.com/MG-RAST/Shock



ENV DIR=/go/src/github.com/MG-RAST/Shock

# set version
RUN cd ${DIR} && \
  VERSION=$(cat VERSION) && \
  sed -i "s/\[% VERSION %\]/${VERSION}/" ${DIR}/shock-server/main.go 


RUN CGO_ENABLED=0 go install -installsuffix cgo -v ...

WORKDIR /go/bin
CMD ["/go/bin/shock-server"]