# Bitcoin Headers Client

[![Release](https://img.shields.io/github/release-pre/libsv/bitcoin-hc.svg?logo=github&style=flat&v=1)](https://github.com/libsv/bitcoin-hc/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/bitcoin-hc/go?logo=github&v=3)](https://github.com/libsv/bitcoin-hc/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/bitcoin-hc?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/bitcoin-hc)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/bitcoin-hc?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

This is a service used to sync and store blockchain headers which can then be used to validate transactions and merkle proofs forming part of an SPV pipeline.

By default, it uses [WhatsOnChain](www.whatsonchain.com) to sync headers using web sockets and stores the data in a local embedded sqlLite database making it highly portable.

## Quick Start

We produce docker images ready for you to use, to get started (within compose) take a look at the [docker-compose](docker-compose.yml) file in the root of the repo.

Using docker is the quickest way to get started.

The only settings you need to be concerned with are:

      - DB_DSN=file:/data/blockheaders.db?_foreign_keys=true&pooled=true;
      - DB_SCHEMA_PATH=/migrations

DSN can be used to change the local database location - this should be a volume mount into the container while SQLite is the only db option, we will support more in future.

DB_SCHEMA_PATH should always be set to /migrations, that's the location within the container where the db migration files are head and will setup the database correctly.

## Endpoints

Currently, we support an HTTP Rest transport, the endpoints are:

### GET http://{{host}}:8442/api/v1/headers/height

Returns the current cached block height as well as the current live network block height.

```json
{
  "height": 696617,
  "networkHeight": 696617,
  "synced": true
}
```

### GET http://{{host}}:8442/api/v1/headers/:blockhash

Returns block information when provided with a blockhash.

```json
{
  "hash": "0000000007e95100bbaf9c467b1416c91ee6c8942d78db630d8d7c4c49eaa717",
  "versionHex": "00000001",
  "merkleroot": "d93d2317ab0a1714e785ca22fe6f906fbafa3242bba251d8bc3a4a057475bec4",
  "bits": "1d00ffff",
  "chainwork": "0000000000000000000000000000000000000000000000000000100110011001",
  "previousblockhash": "0000000066ca066a388fea7b34b7ff1e0e6f87f97be2a1eb82ed574182664fd4",
  "nextblockhash": "00000000860819695ff009f818bbceefa24da888aad021e8a2701c614fa5686c",
  "confirmations": 689958,
  "height": 4096,
  "mediantime": 1234507372,
  "difficulty": 1,
  "version": 1,
  "time": 1234509841,
  "nonce": 3594350374
}
```