# Description
A Go programming language implementation of the Raft distributed consensus algorithm. Raft is a protocol for managing the data of replica servers, if a replica server fails but later recovers, Raft brings the replica server’s data up to date.

## Run Instructions
* Ensure Docker is installed on your PC.
* Execute these commands in the root directory of the cloned repo:

Setting up Docker container:

```
docker build -t raft .
docker run --name=raft_cont -td raft
docker exec -it raft_cont bash
```
A bash terminal running ubuntu should appear

Running raft tests:
* In the terminal that appears, execute these commands:
```
cd 486; export GOPATH="$PWD"
cd src; go work init
cd raft; go mod init
cd ..; go work use -r .
cd "$GOPATH/src/raft"
go test -run Election
```


