[![Release](https://img.shields.io/github/release-pre/bitcoin-sv/block-headers-service.svg?logo=github&style=flat&v=1)](https://github.com/bitcoin-sv/block-headers-service/releases)
[![Build Status](https://img.shields.io/github/workflow/status/bitcoin-sv/block-headers-service/go?logo=github&v=3)](https://github.com/bitcoin-sv/block-headers-service/actions)
[![Report](https://goreportcard.com/badge/github.com/bitcoin-sv/block-headers-service?style=flat&v=1)](https://goreportcard.com/report/github.com/bitcoin-sv/block-headers-service)
[![Go](https://img.shields.io/github/go-mod/go-version/bitcoin-sv/block-headers-service?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-bitcoin-sv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/bitcoin-sv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)
<br />

<h1 id="top" align="center">Block Headers Service</h1>
<h3 id="top" align="center">formerly Pulse</h1>

  <p align="center">
    Go application used to collect and return information about blockchain headers
</p>

## Table of contents
<details>
  <!--<summary>Table of Contents</summary> -->
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#main-functionality">Main Functionality</a></li>
      </ul>
    </li>
<li>
      <a href="#how-to-use-it">How to use it</a>
<ul>
        <li><a href="#docker-image">Docker image</a></li>
      </ul>
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
    <li>
<a href="#Running-from-source">Running from source</a>
<ul>
        <li><a href="#Running-from-source">Run application</a></li>
      </ul>
      <ul>
        <li><a href="#configuration">Configuration</a></li>
      </ul>
      
    </li>
    
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

Block header service (formerly named "Pulse") is a go service which connects into BSV P2P network to gather and then serve information about all exisiting and new headers. It is build to work as a standaolne app or a module in bigger system.

#### Main functionality
The main functionality of the application is synchornization with peers and collecting all headers. After starting the server, it creates default objects and connects to BSV P2P network. Application has defined checkpoints (specific headers) which are used in synchronization. During this process, server is asking peers for headers (from checkpoint to checkpoint) in batches of 2000. Every header received from peers is saved in memory. After full synchronization, server is changing the operating mode and start to listening for new header. After when new block has been mined, this information should be sended from peers to our server.

## How to use it

### Docker image

Pull image from docker hub https://hub.docker.com/r/bsvb/block-headers-service
1. ```docker pull bsvb/block-headers-service```

Starting new instance
1. ```docker run bsvb/block-headers-service:latest```


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
To do it, just add the following header to all the requests to block-headers-service.
```
Authorization Bearer replace_me_with_token_you_want_to_use_as_admin_token
```

#### Additional tokens

If you have a need for additional tokens to authenticate in block-headers-service.
you can generate such with the following request:
```http request
POST https://{{block-headers-service_url}}/api/v1/access
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
DELETE https://{{block-headers-service_url}}/api/v1/access/{{some_token_created_by_server}}
Authorization: Bearer replace_me_with_token_you_want_to_use_as_admin_token
```
After this request succeeded the token can't be used to authenticate in block-headers-service.

### Websocket

Block headers service can notify a client via websockets that new header was received and store by it.

#### Subscribing

Block headers service use [centrifugal/centrifuge](https://github.com/centrifugal/centrifuge) to run a server.
Therefore, to integrate you need to choose a client library matching a programming language of your choice.

Example how to subscribe using GO lang library [centrifugal/centrifuge-go](https://github.com/centrifugal/centrifuge-go) 
can be found in [./examples/ws-subscribe-to-new-headers/](./examples/ws-subscribe-to-new-headers/main.go)

### Webhooks

#### Creating webhook
Creating a new webhook is done via POST request
```http request
 POST https://{{block-headers-service_url}}/api/v1/webhook
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
 GET https://{{block-headers-service_url}}/api/v1/webhook?url={{webhook_url}}
 ```

#### Revoke webhook
If you want to revoke webhook you can use the following request:
```http request
 DELETE https://{{block-headers-service_url}}/api/v1/webhook?url={{webhook_url}}
 ```
This request will delete webhook permanently

#### Refresh webhook
If the number of failed requests wil exceed `WEBHOOK_MAXTRIES`, webhook will be set to inactive. To refresh webhook you can use this same endpoint as for webhook creation.

### Running from source

1. Install Go according to the installation instructions here: http://golang.org/doc/install

Options to run Block Headers Service:

a) Clone the repo

   ```sh
  git clone https://github.com/bitcoin-sv/block-headers-service
   ``` 
1. ```go run ./cmd/main.go```

Or run app with docker
1. ```docker compose up --build```

b) Get package from ``pkg.dev.go``
1. ```go get -u https://pkg.go.dev/github.com/bitcoin-sv/block-headers-service```
2. ```go build -o block-headers-service```
3. ```./block-headers-service```


## Usage

> Every variable which is used and can be configured is described in [config.example.yaml](config.example.yaml)

### Defaults

If you run block headers service without editing anything, it will use the default configuration from file [defaults.go](/config/defaults.go). It is set up to use _sqlite_ database with enabled authorization (with default auth key) for http server.

### Config Variables

Default config variables can be overridden by (in this order of importance):

1. Flags (only the ones below)
2. ENV variables
3. Config file

#### Flags

Available flags:

```bash
  -C, --config_file string                       custom config file path
  -h, --help                                     show help
  -v, --version                                  show version
  -d, --dump_config                              dump config to file, specified by config_file (-C) flag
  -e, --export_headers                           export headers to file
```

To generate config file with defaults, use the --dump flag, or:

```bash
go run ./cmd/main.go -d
```

The default config file path is **program root**, and the default file name is **config.yaml**. This can be overridden by -C flag.

```bash
go run ./cmd/main.go -C /my/config.yaml
```

#### Environment variables

To override any config variable with ENV, use the "headers-service\_" prefix with mapstructure annotation path with "_" as a delimiter in all uppercase. Example:

Let's take this fragment of AppConfig from `config.example.yaml`:

```yaml
websocket:
  # Maximum number of history items
  history_max: 300
  # History time-to-live
  history_ttl: 10
```

To override history_max in websocket config, use the path with "_" as a path delimiter and bhs\_ as prefix. So:

```bash
BHS_HISTORY_MAX=300
```


<!-- PROJECT LOGO -->
<br />
