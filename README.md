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
        <li><a href="#authentication">Authentication</a></li>
      </ul>
      <ul>
        <li><a href="#websocket">Websocket</a></li>
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
Configuration processing is carried out by the viper.
There are 3 ways how default configuration can be overridden:
1. Flags
2. Env variables
3. Сonfig file
* Note that the rewrite methods are listed from highest to lowest priority. This means that if a parameter is passed as a flag and as an environment variable and is present in the configuration file, then the value of the flag will be used by the viper since the flag has the highest priority.
* By default pulse is not reading any file config. The file will be read only in case if -C or --p2pconfig flag with config path is passed.


Default P2P config:
```go
DefaultLogLevel                = "info"     // Options: trace, debug, info, warn, error, critical
DefaultLogDirname              = "logs"     // Directory to log output
DefaultLogFilename             = "p2p.log"  // File to log output
DefaultMaxPeers                = 125        // Max number of inbound and outbound peers
DefaultMaxPeersPerIP           = 5          // Max number of inbound and outbound peers per IP
DefaultBanDuration             = time.Hour * 24 // How long to ban misbehaving peers. Minimum 1 second
DefaultConnectTimeout          = time.Second * 30   // How long to wait for peer connection
DefaultTrickleInterval         = peer.DefaultTrickleInterval // Minimum time between attempts to send new inventory to a connected peer
DefaultExcessiveBlockSize      = 128000000  // The maximum size block (in bytes) this node will accept. Cannot be less than 32000000.
DefaultMinSyncPeerNetworkSpeed = 51200      // Disconnect sync peers slower than this threshold in bytes/sec
DefaultTargetOutboundPeers     = uint32(8)  // Number of outbound connections to maintain
DefaultBlocksToConfirmFork     = 10         // Number of blocks that confirm the right chain fork
```

Settings related to database:

      - DB_DSN=file:/data/blockheaders.db?_foreign_keys=true&pooled=true
      - DB_SCHEMAPATH=/migrations
      - DB_PREPAREDDB=true
      - DB_PREPAREDDBFILEPATH="./data/blockheaders.xz"

DSN can be used to change the local database location - this should be a volume mount into the container while SQLite is the only db option, we will support more in future.

DB_SCHEMA_PATH should always be set to /migrations, that's the location within the container where the db migration files are head and will setup the database correctly.

DB_PREPAREDDB is used to define if application should use prepared db.

DB_PREPAREDDBFILEPATH define path to prepared db.


Settings related to admin auth:
      - HTTP_AUTHTOKEN=admin_only_afUMlv5iiDgQtj22O9n5fADeSb

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

`HTTP_AUTHTOKEN=replace_me_with_token_you_want_to_use_as_admin_token`  

#### Disabling Auth Requirement  

To disable authentication exposing all endpoints openly, set the following environment variable. 
This is available if you prefer to use your own authentication in a separate proxy or similar. 
We do not recommend you expose the server to the internet without authentication, 
as it would then be possible for anyone to prune your headers at will.  

`HTTP_USEAUTH=false`  

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

### Websocket

Pulse can notify a client via websockets that new header was received and store by it.

#### Subscribing

Pulse use [centrifugal/centrifuge](https://github.com/centrifugal/centrifuge) to run a server. 
Therefore, to integrate you need to choose a client library matching a programming language of your choice.

Example how to subscribe using GO lang library [centrifugal/centrifuge-go](https://github.com/centrifugal/centrifuge-go) 
can be found in [./examples/ws-subscribe-to-new-headers/](./examples/ws-subscribe-to-new-headers/main.go)

### Webhooks

#### Creating webhook
Creating a new webhook is done via POST request
```http request
 POST https://{{pulse_url}}/api/v1/webhook
 ```

 Data which should be sent in body:
 ```
{
  "url": "<server_url>",
  "requiredAuth": {
    "type": "BEARER|CUSTOM_HEADER",
    "token": "<authorization_token>",
    "header": "<custom_header_name>",      
  }
}
 ```

 Information:
  - If authorization is enabled this request also requires `Authorization` header
  - url have to include http or https protocol example: `https://test-url.com`
  - requiredAuth is used to define authorization for webhook
    - type `BEARER` - token will be placed in `Authorization: Bearer {{token}}` header
    - type `CUSTOM_HEADER`  - authorization header will be build from given variables `{{header}}: {{token}}`

Example response:
````json
{
  "url": "http://example.com/api/v1/webhook/new-header",
  "createdAt": "2023-05-11T13:05:23.297808+02:00",
  "lastEmitStatus": "",
  "lastEmitTimestamp": "0001-01-01T00:00:00Z",
  "errorsCount": 0,
  "active": true
}
````
After that webhook is created and will be informed about new headers.

#### Check webhook
To check webhook you can use the GET request which will return webhook object (same as when creating new webhook) from which you can get all the information
```http request
 GET https://{{pulse_url}}/api/v1/webhook?url={{webhook_url}}
 ```

#### Revoke webhook
If you want to revoke webhook you can use the following request:
```http request
 DELETE https://{{pulse_url}}/api/v1/webhook?url={{webhook_url}}
 ```
This request will delete webhook permanently

#### Refresh webhook
If the number of failed requests wil exceed `WEBHOOK_MAXTRIES`, webhook will be set to inactive. To refresh webhook you can use this same endpoint as for webhook creation.

<!-- PROJECT LOGO -->
<br />
