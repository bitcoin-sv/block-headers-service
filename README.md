
# Bitcoin Headers Client

[![Release](https://img.shields.io/github/release-pre/libsv/pulse.svg?logo=github&style=flat&v=1)](https://github.com/libsv/pulse/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/pulse/go?logo=github&v=3)](https://github.com/libsv/pulse/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/pulse?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/pulse)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/pulse?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)
<br />

<h1 id="top" align="center">Pulse</h1>

  <p align="center">
    Go application used to collect and return information about blockchain headers
</div>

## Table of contents
<details>
  <!--<summary>Table of Contents</summary> -->
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
      <ul>
        <li><a href="#main-functionality">Main Functionality</a></li>
      </ul>
      <ul>
        <li><a href="#current-database-structures">Current database structures</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#installation">Installation</a></li>
      </ul>
      <ul>
        <li><a href="#configuration">Configuration</a></li>
      </ul>
      <ul>
        <li><a href="#run-application">Run application</a></li>
      </ul>
    </li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project


Pulse is a go server which connects into BSV P2P network to gather and then serve information about all exisiting and new headers. It is build to work as a standaolne app or a module in bigger system.

#### Main functionality
The main functionality of the application is synchornization with peers and collecting all headers. After starting the server, it creates default objects and connects to BSV P2P network. Application has defined checkpoints (specific headers) which are used in synchronization. During this process, server is asking peers for headers (from checkpoint to checkpoint) in batches of 2000. Every header received from peers is saved in memory. After full synchronization, server is changing the operating mode and start to listening for new header. After when new block has been mined, this information should be sended from peers to our server.

#### Current database structures
We store everything in database and then we calculate any important information related to: confirmations, state etc.
The only exception is when inserting a new header we check if he has any parent in database and if we don't have any parents 
we set the "isorphan" flag to true.
Database schema below:

![Alt text](docs/images/headers.png?raw=true "Headers database structure")
<p align="center"><i>Headers table schema with sqlite database</i></p>


#### Built With

* Go 1.19

<!-- GETTING STARTED -->
## Getting Started

### Installation
1. Install Go according to the installation instructions here: http://golang.org/doc/install
2. Clone the repo
   ```sh
   https://github.com/gignative-solutions/ba-p2p-headers.git
   ```
    
### Configuration
In the ```config.go``` is the configuration of the application. By changing variables you can adjust the work of our server

```
defaultConfigFilename          = "p2p.conf"
defaultLogLevel                = "info"
defaultLogDirname              = "logs"
defaultLogFilename             = "p2p.log"
defaultMaxPeers                = 125
defaultMaxPeersPerIP           = 5
defaultBanDuration             = time.Hour * 24
defaultConnectTimeout          = time.Second * 30
defaultTrickleInterval         = peer.DefaultTrickleInterval
defaultExcessiveBlockSize      = 128000000
defaultMinSyncPeerNetworkSpeed = 51200
defaultTargetOutboundPeers     = uint32(8)
defaultBlocksToConfirmFork     = 10
```

Settings related to database:

      - DB_DSN=file:/data/blockheaders.db?_foreign_keys=true&pooled=true
      - DB_SCHEMA_PATH=/migrations

DSN can be used to change the local database location - this should be a volume mount into the container while SQLite is the only db option, we will support more in future.

DB_SCHEMA_PATH should always be set to /migrations, that's the location within the container where the db migration files are head and will setup the database correctly.

## 

### Run application
```sh
go run ./cmd/ .
```
or with Docker
```sh
docker compose up --build
```

### Endpoints documentation
For endpoints documentation you can visit swagger which is exposed on port 8080 by default.
```
http://localhost:8080/swagger/index.html
```

<!-- PROJECT LOGO -->
<br />
