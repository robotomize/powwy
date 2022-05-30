# Powwy

[![codebeat badge](https://codebeat.co/badges/a4a12b24-98e6-4627-b01c-8b124561f2e1)](https://codebeat.co/projects/github-com-robotomize-powwy-main)
[![Build status](https://github.com/robotomize/powwy/actions/workflows/release.yml/badge.svg)](https://github.com/robotomize/powwy/actions)
[![GitHub license](https://img.shields.io/github/license/robotomize/powwy.svg)](https://github.com/robotomize/powwy/blob/master/LICENSE)

A toy implementation of a tcp server with its own simple text protocol with content protection from ddos attacks using
proof of work.

The project consists of a tcp server and a tcp client.
The server offers to solve the proof of work problem to the client in order to access the content.  
The client receives the problem, solves it, and receives a quote from Words of wisdom.

## Requirements

* go1.18
* golangci-lint

## Usage

Install from local go toolkit [Go 1.18](https://go.dev/dl/)

### Go install
```sh
go install ./cmd/powwy-srv
go install ./cmd/powwy-cli
    
```

### Make build

```sh
make build
```

### Docker install

```sh
sudo docker build -it powwy-cli -f ./docker/client.Dockerfile .
sudo docker build -it powwy -f ./docker/server.Dockerfile .

sudo docker run powwy --net=host -p 3333:3333
sudo docker run powwy-cli -a localhost:3333 -d // infinite loop

sudo docker run powwy-cli compute <header>
```

From docker hub
```sh
docker pull robotomize/powwy-cli:latest
docker pull robotomize/powwy-srv:latest
```

### Docker compose

```sh
sudo docker-compose up
```

## CLI powwy

Find a solution to the HashCash header problem

```sh
powwy-cli compute <header1> <header2> 
```

```sh
Usage: powwy-cli compute <header> <header> <header>

Usage:

compute [flags]

Flags:
-h, --help   help for compute

Global Flags:
-d, --duration duration   -d 10s (default -1ns)
-i, --iterations int      -i 1000000 (default -1)
-w, --workers int         -w 4 (default 2)
```

Connect to powwy server and fetch quotes

```sh
powwly-cli -a localhost:3333
```

```sh
Usage:

Available Commands:
completion  Generate the autocompletion script for the specified shell
compute     Usage: powwy-cli compute <header> <header> <header>
help        Help about any command

Flags:
-a, --addr string         -a localhost:3333 (default "localhost:3333")
-s, --dos                 -s true
-d, --duration duration   -d 10s (default -1ns)
-h, --help                help for this command
-i, --iterations int      -i 1000000 (default -1)
-n, --network string      -n tcp4 (default "tcp")
-w, --workers int         -w 4 (default 2)

Use " [command] --help" for more information about a command.
```

## Proof of work

### HashCash

The [http://www.hashcash.org/](HashCash) algorithm was chosen to implement the proof of work mechanism

Pros:
* Easy to find description
* Easy to implement

Cons:
* Difficult to adjust the difficulty of the task

### Implementing the header

*HashCashHeader: version:difficult:expiredAt:subject:alg:nonce:counter*

* version - version of the header, represented by an int
* difficult - the complexity of the useful work, expressed in the number of first zeros of the hash
* expiredAt - the lifetime of the task, after which it is considered invalid
* subject - defines the name of the resource or other identifying information, such as a user ID or its ip address
* alg - determines what type of hash to use
* nonce - is a randomly generated set of bytes
* counter - Current counter value

*Example 1:20:1665396610:localhost:sha-512:hVscDCMZcS1WYg==:BQAAAAAAAAA=*

### Merkle tree

[Merkle in ethereum](https://blog.ethereum.org/2015/11/15/merkling-in-ethereum/)

Pros:
  Works great for decentralized applications

Cons:
  Not very suitable for the implementation of client-server applications

## Protocol

powwy uses a simple command text protocol

List of available commands

```go
// Proto specification tags
const (
    REQ = "REQ" // REQ - request challenge
    RES = "RES" // RES - request resource
    RSV  = "RSV" // RSV - response with payload
    OK = "OK"    // OK - command accepted
    ERR = "ERR"   // ERR - command err
    DISC = "DISC" // DISC - initialize close connection
)

```

Format:
`<Command> <body length> |<body>`
or
`<Command>`

Example:

```sh
---------------------------------------------------------------------------------------------------------------
UNKNOWN_COMMAND
ERR 18 |command wrong
---------------------------------------------------------------------------------------------------------------
REQ
ERR 21 |internal server error
---------------------------------------------------------------------------------------------------------------
REQ
RSR 62 |1:5:1665396610:localhost:sha-256:vZOxuoIgixP+hw==:AAAAAAAAAAA=
---------------------------------------------------------------------------------------------------------------
RES 62 |1:5:1665396610:localhost:sha-256:vZOxuoIgixP+hw==:FxUQAAAAAAA=
RST 139 |Voice is not just the sound that comes from your throat, but the feelings that come from your words.
― Jennifer Donnelly, A Northern Light
---------------------------------------------------------------------------------------------------------------
DISC
OK
---------------------------------------------------------------------------------------------------------------
```





