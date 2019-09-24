##### This test container can be ran in skyport2 enviroment to test shock-server
#
# docker build -t mgrast/shock-test -f test/Dockerfile .
# docker run --rm --name test -t -d mgrast/shock-test ash
# docker exec test ash -c 'HOSTIP=`/sbin/ip route | head -1 | cut -f3 -d" "`; echo "$HOSTIP skyport.local" >> /etc/hosts'
# docker exec test /shock-tester.sh -h http://skyport.local -p 7445
#
# Note the setcap Linux command will only succeed if run as root.
# This allows the shock-server to bind to port 80 if desired.
#setcap 'cap_net_bind_service=+ep' bin/shock-server

FROM golang:alpine

RUN apk update && apk add \
    git \
    python3

RUN pip3 install \
    pytest \
    requests 

ENV DIR=/go/src/github.com/MG-RAST/Shock
WORKDIR /

COPY . /go/src/github.com/MG-RAST/Shock

RUN mkdir -p /data/db /var/log/shock /usr/local/shock/data ${DIR} && \
    ln -s ${DIR}/test/testdata /testdata && \
    ln -s ${DIR}/test/shock-tester.sh /shock-tester.sh

# compile
RUN cd ${DIR} && \
      CGO_ENABLED=0 go install -a -installsuffix cgo -v ./shock-client/

# run tests
CMD ["/bin/ash"]