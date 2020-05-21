

# docker build -t mgrast/shock .
# docker run --rm --name test -ti mgrast/shock /bin/ash

# Note the setcap Linux command will only succeed if run as root.
# This allows the shock-server to bind to port 80 if desired.
#setcap 'cap_net_bind_service=+ep' bin/shock-server

FROM golang:alpine

ENV PYTHONUNBUFFERED=1

RUN apk update && apk add git curl &&\
    echo "**** install Python ****" && \
    apk add --no-cache python3 && \
    if [ ! -e /usr/bin/python ]; then ln -sf python3 /usr/bin/python ; fi && \
    \
    echo "**** install pip ****" && \
    python3 -m ensurepip && \
    rm -r /usr/lib/python*/ensurepip && \
    pip3 install --no-cache --upgrade pip setuptools wheel && \
    if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip ; fi &&\
    pip3 install boto3

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


CMD ["/go/bin/shock-server"]
