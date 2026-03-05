

# docker build -t mgrast/shock .
# docker run --rm --name test -ti mgrast/shock /bin/ash

# Note the setcap Linux command will only succeed if run as root.
# This allows the shock-server to bind to port 80 if desired.
#setcap 'cap_net_bind_service=+ep' bin/shock-server

FROM node:20-alpine AS ui-builder
WORKDIR /build/clients
COPY clients/ .
RUN npm install && \
    cd shock-ts && npm run build && \
    cd ../shock-ui && npm run build

FROM golang:alpine

ENV PYTHONUNBUFFERED=1

RUN apk update && apk add --no-cache git curl python3 py3-pip py3-boto3 && \
    if [ ! -e /usr/bin/python ]; then ln -sf python3 /usr/bin/python ; fi

ENV DIR=/go/src/github.com/MG-RAST/Shock
WORKDIR /go/bin

COPY . /go/src/github.com/MG-RAST/Shock

# Copy built UI into the Go embed directory
COPY --from=ui-builder /build/clients/shock-ui/dist ${DIR}/shock-server/ui/dist

RUN mkdir -p /var/log/shock /usr/local/shock/data ${DIR}

# set version
#RUN cd ${DIR} && \
#  VERSION=$(cat VERSION) && \
#  sed -i "s/\[% VERSION %\]/${VERSION}/" shock-server/conf/conf.go

# compile (uses vendored dependencies via -mod=vendor)
RUN cd ${DIR} &&\
     ./compile-server.sh

RUN mkdir -p /etc/shock.d/ /var/cache/shock && touch /etc/shock.d/shock-server.conf

# Put boto3 S3 scripts on PATH so exec.Command can find them by bare name
RUN cp ${DIR}/shock-server/plug-ins/boto-s3-*.py /usr/local/bin/ && \
    chmod +x /usr/local/bin/boto-s3-*.py

CMD ["/go/bin/shock-server"]
