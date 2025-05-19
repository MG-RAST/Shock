FROM ubuntu:latest

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y  && apt-get upgrade -y

RUN apt-get -y install \
    curl \
    cwltool \
    less \
    python3 \
    python3-boto3 \
    python3-pip \
    python3-requests \
    vim

# Docker 
RUN apt-get -y install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN apt-key fingerprint 0EBFCD88
RUN add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
RUN apt-get install -y  docker-ce docker-ce-cli containerd.io


# RUN update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.6 1 && \
#     update-alternatives --install /usr/bin/python python /usr/bin/python3.6 2
RUN  update-alternatives --install /usr/bin/python python /usr/bin/python3 1
COPY *.py /usr/local/bin/
RUN chmod a+x /usr/local/bin/*

