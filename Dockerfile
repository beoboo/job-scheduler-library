FROM ubuntu:20.04

ENV GOLANG_VERSION 1.17.3

# Install packages
RUN apt update && \
    apt install -y build-essential wget &&  \
    rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN rm -f go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH "$PATH:/usr/local/go/bin"

ADD . /src

WORKDIR /src

ENTRYPOINT "/bin/bash"