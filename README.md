# Bitcoin Headers Client

[![Release](https://img.shields.io/github/release-pre/libsv/bitcoin-hc.svg?logo=github&style=flat&v=1)](https://github.com/libsv/bitcoin-hc/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/bitcoin-hc/go?logo=github&v=3)](https://github.com/libsv/bitcoin-hc/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/bitcoin-hc?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/bitcoin-hc)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/bitcoin-hc?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

This is a service used to sync and store blockchain headers which can then be used to validate transactions and merkle proofs forming part of an SPV pipeline.

By default, it uses [WhatsOnChain](www.whatsonchain.com) to sync headers using web sockets and stores the data in a local embedded sqlLite database making it highly portable.

