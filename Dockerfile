FROM ubuntu:latest

# prerequisites
RUN apt-get update --fix-missing
RUN apt-get update && apt-get install -y apt-utils
RUN apt-get install -y gcc make build-essential libcunit1-dev libcunit1-doc libcunit1 wget python3 qemu xorg-dev libncurses5-dev gdb git

# copying local files to docker container
COPY /raft /home/486


# CSE 486 Distributed Systems
RUN wget -P /tmp https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# extract tar
RUN tar -C /usr/local -xzf /tmp/go1.21.6.linux-amd64.tar.gz
RUN rm /tmp/go1.21.6.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR /home