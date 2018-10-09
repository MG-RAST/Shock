

# docker build -t mgrast/shock .
# docker run --rm --name test -ti mgrast/shock /bin/ash

# Note the setcap Linux command will only succeed if run as root.
# This allows the shock-server to bind to port 80 if desired.
#setcap 'cap_net_bind_service=+ep' bin/shock-server

FROM golang:1.7.6-alpine

RUN apk update && apk add git mongodb

ENV DIR=/go/src/github.com/MG-RAST/Shock
WORKDIR /go/bin

COPY . /go/src/github.com/MG-RAST/Shock

RUN mkdir -p /data/db /var/log/shock /usr/local/shock/data ${DIR} && \
    ln -s ${DIR}/test/testdata /testdata && \
    ln -s ${DIR}/test/shock-tester.sh /shock-tester.sh && \
    ln -s ${DIR}/shock-server.sh /etc/init.d/shock-server.sh && \
    ln -s ${DIR}/shock-server.conf.template /usr/local/shock/shock-server.conf

# set version
RUN cd ${DIR} && \
  VERSION=$(cat VERSION) && \
  sed -i "s/\[% VERSION %\]/${VERSION}/" shock-server/conf/conf.go

# compile
RUN cd ${DIR} && \
    go get -d ./shock-server/ ./shock-client/  && \
    CGO_ENABLED=0 go install -a -installsuffix cgo -v ./shock-server/ ./shock-client/

# run tests
CMD ["/shock-tester.sh"]