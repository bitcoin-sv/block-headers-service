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
    <li>
      <a href="#how-to-use-it">How to use it</a>
      <ul>
        <li><a href="#endpoints-documentation">Endpoints documentation</a></li>
      </ul>
      <ul>
        <li><a href="#webhooks">Webhooks</a></li>
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
      - DB_PREPAREDDB=true
      - DB_PREPAREDDBFILE_PATH="./data/blockheaders.xz"

DSN can be used to change the local database location - this should be a volume mount into the container while SQLite is the only db option, we will support more in future.

DB_SCHEMA_PATH should always be set to /migrations, that's the location within the container where the db migration files are head and will setup the database correctly.

DB_PREPAREDDB is used to define if application should use prepared db.

DB_PREPAREDDBFILE_PATH define path to prepared db.


Settings related to admin auth:
      - HTTP_SERVER_AUTHTOKEN=admin_only_afUMlv5iiDgQtj22O9n5fADeSb

This admin token should be used as a Bearer token in the Authorization header when dynamically creating secure tokens for applications to then use at the POST /api/v1/access endpoint.  

## 

### Run application
```sh
go run ./cmd/ .
```
or with Docker
```sh
docker compose up --build
```

## How to use it

### Endpoints documentation
For endpoints documentation you can visit swagger which is exposed on port 8080 by default.
```
http://localhost:8080/swagger/index.html
```

### Authentication

#### Enabled by Default

The default assumes you want to use Authentication. This requires a single environment variable.  

`HTTP_SERVER_AUTHTOKEN=replace_me_with_token_you_want_to_use_as_admin_token`  

#### Disabling Auth Requirement  

To disable authentication exposing all endpoints openly, set the following environment variable. 
This is available if you prefer to use your own authentication in a separate proxy or similar. 
We do not recommend you expose the server to the internet without authentication, 
as it would then be possible for anyone to prune your headers at will.  

`HTTP_SERVER_USEAUTH=false`  

#### Authenticate with admin token

After the setup of authentication you can use provided token to authenticate.
To do it, just add the following header to all the requests to pulse
```
Authorization Bearer replace_me_with_token_you_want_to_use_as_admin_token
```

#### Additional tokens

If you have a need for additional tokens to authenticate in pulse 
you can generate such with the following request:
```http request
POST https://{{pulse_url}}/api/v1/access
Authorization: Bearer replace_me_with_token_you_want_to_use_as_admin_token
```
In response you should receive something like
```json
{
  "token": "some_token_created_by_server",
  "createdAt": "2023-05-11T10:20:16.227582Z",
  "isAdmin": false
}
```
Now you can put a value from "token" property from the response and use it in all requests to server by setting header:
```http header
Authorization: Bearer some_token_created_by_server
```

If at some point you want to revoke this additional token you can make a request:
```http request
DELETE https://{{pulse_url}}/api/v1/access/{{some_token_created_by_server}}
Authorization: Bearer replace_me_with_token_you_want_to_use_as_admin_token
```
After this request succeeded the token can't be used to authenticate in pulse.

<!-- PROJECT LOGO -->
<br />
