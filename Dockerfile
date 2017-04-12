

# docker build -t mgrast/shock .
# docker run --rm --name test -ti mgrast/shock /bin/ash

FROM golang:1.7.5-alpine

ENV DIR=/go/src/github.com/MG-RAST/Shock
WORKDIR /go/bin

COPY . /go/src/github.com/MG-RAST/Shock

RUN mkdir -p /var/log/shock /usr/local/shock ${DIR}

# set version
RUN cd ${DIR} && \
  VERSION=$(cat VERSION) && \
  sed -i "s/\[% VERSION %\]/${VERSION}/" shock-server/main.go 

# compile
RUN cd ${DIR} && \
    go get -d ./shock-server/ && \
    CGO_ENABLED=0 go install -a -installsuffix cgo -v ./shock-server/

CMD ["/go/bin/shock-server"]