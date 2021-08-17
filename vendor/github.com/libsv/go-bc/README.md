# go-bc

> The go-to Bitcoin Block Chain (BC) GoLang library  

[![Release](https://img.shields.io/github/release-pre/libsv/go-bc.svg?logo=github&style=flat&v=1)](https://github.com/libsv/go-bc/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/go-bc/run-go-tests?logo=github&v=3)](https://github.com/libsv/go-bc/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/go-bc?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/go-bc)
[![codecov](https://codecov.io/gh/libsv/go-bc/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/libsv/go-bc)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/go-bc?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

<br/>

## Table of Contents

- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

<br/>

## Installation

**go-bc** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

```shell script
go get -u github.com/libsv/go-bc
```

<br/>

## Documentation

View the generated [documentation](https://pkg.go.dev/github.com/libsv/go-bc)

[![GoDoc](https://godoc.org/github.com/libsv/go-bc?status.svg&style=flat)](https://pkg.go.dev/github.com/libsv/go-bc)

For more information around the technical aspects of Bitcoin, please see the updated [Bitcoin Wiki](https://wiki.bitcoinsv.io/index.php/Main_Page)

<br/>

### Features

- Block header building
- Coinbase transaction building (cb1 + cb2 in stratum protocol)
- Bitcoin block hash difficulty and hashrate functions
- Merkle proof/root/branch functions

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands

```shell script
make help
```

List of all current commands:

```text
all                  Runs multiple commands
clean                Remove previous builds and any test cache data
clean-mods           Remove all the Go mod cache
coverage             Shows the test coverage
godocs               Sync the latest tag with GoDocs
help                 Show this help message
install              Install the application
install-go           Install the application (Using Native Go)
lint                 Run the golangci-lint application (install if not found)
release              Full production release (creates release in Github)
release              Runs common.release then runs godocs
release-snap         Test the full release (build binaries)
release-test         Full production test release (everything except deploy)
replace-version      Replaces the version in HTML/JS (pre-deploy)
tag                  Generate a new tag and push (tag version=0.0.0)
tag-remove           Remove a tag if found (tag-remove version=0.0.0)
tag-update           Update an existing tag to current commit (tag-update version=0.0.0)
test                 Runs vet, lint and ALL tests
test-ci              Runs all tests via CI (exports coverage)
test-ci-no-race      Runs all tests via CI (no race) (exports coverage)
test-ci-short        Runs unit tests via CI (exports coverage)
test-short           Runs vet, lint and tests (excludes integration tests)
uninstall            Uninstall the application (and remove files)
update-linter        Update the golangci-lint package (macOS only)
vet                  Run the Go vet application
```

</details>

<br/>

## Examples & Tests

All unit tests and [examples](examples) run via [Github Actions](https://github.com/libsv/go-bc/actions) and
uses [Go version 1.15.x](https://golang.org/doc/go1.15). View the [configuration file](.github/workflows/run-tests.yml).

Run all tests (including integration tests)

```shell script
make test
```

Run tests (excluding integration tests)

```shell script
make test-short
```

<br/>

## Benchmarks

Run the Go benchmarks:

```shell script
make bench
```

<br/>

## Code Standards

Read more about this Go project's [code standards](CODE_STANDARDS.md).

<br/>

## Usage

View the [examples](examples)

<br/>

## Maintainers

| [<img src="https://github.com/jadwahab.png" height="50" alt="JW" />](https://github.com/jadwahab)  |
|:---:|
|  [JW](https://github.com/jadwahab) |

<br/>

## Contributors

| [<img src="https://github.com/jadwahab.png" height="50" alt="JW" />](https://github.com/jadwahab) | [<img src="https://github.com/ordishs.png" height="50" alt="SO" />](https://github.com/ordishs) | [<img src="https://avatars.githubusercontent.com/u/26772?s=400&v=4" height="50" alt="LM" />](https://github.com/liam) |
|:---:|:---:|:---:|
| [JW](https://github.com/jadwahab) | [SO](https://github.com/ordishs) | [LM](https://github.com/liam) |

<br/>

## Contributing

View the [contributing guidelines](CONTRIBUTING.md) and please follow the [code of conduct](CODE_OF_CONDUCT.md).

### How can I help?

All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/libsv) :clap:
or by making a [**bitcoin donation**](https://gobitcoinsv.com/#sponsor) to ensure this journey continues indefinitely! :rocket:

<br/>

## License

![License](https://img.shields.io/github/license/libsv/go-bc.svg?style=flat&v=1)
