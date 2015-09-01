
# creates statically compiled shock-server binary: /gopath/bin/shock-server

FROM ubuntu:14.10

RUN apt-get update && apt-get install -y \
  make \
  git-core \
  curl

RUN curl -s https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -v -C /usr/local -xz

ENV GOROOT /usr/local/go 
ENV PATH /usr/local/go/bin:/gopath/bin:/usr/local/bin:$PATH 
ENV GOPATH /gopath/

RUN mkdir -p /gopath/ && \
  cd /gopath/ && \
  curl -s -O https://raw.githubusercontent.com/MG-RAST/Shock/master/Makefile && \
  make install
