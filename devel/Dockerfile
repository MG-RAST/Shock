

# docker build -t mgrast/shock .
# docker run --rm --name test -ti mgrast/shock /bin/ash

# Note the setcap Linux command will only succeed if run as root.
# This allows the shock-server to bind to port 80 if desired.
#setcap 'cap_net_bind_service=+ep' bin/shock-server

FROM golang:alpine

RUN apk update && apk add git curl

ENV DIR=/go/src/github.com/MG-RAST/Shock
WORKDIR /go/bin

COPY . /go/src/github.com/MG-RAST/Shock

RUN mkdir -p /var/log/shock /usr/local/shock/data ${DIR}

# set version
#RUN cd ${DIR} && \
#  VERSION=$(cat VERSION) && \
#  sed -i "s/\[% VERSION %\]/${VERSION}/" shock-server/conf/conf.go

# compile
RUN cd ${DIR} && \
     go get github.com/MG-RAST/go-shock-client  &&\
     ./compile-server.sh


CMD ["/bin/sleep 9999d"]
