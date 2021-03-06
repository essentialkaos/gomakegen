<p align="center"><a href="#readme"><img src="https://gh.kaos.st/gomakegen.svg"/></a></p>

<p align="center">
  <a href="https://github.com/essentialkaos/gomakegen/actions"><img src="https://github.com/essentialkaos/gomakegen/workflows/CI/badge.svg" alt="GitHub Actions Status" /></a>
  <a href="https://github.com/essentialkaos/gomakegen/actions?query=workflow%3ACodeQL"><img src="https://github.com/essentialkaos/gomakegen/workflows/CodeQL/badge.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/gomakegen"><img src="https://goreportcard.com/badge/github.com/essentialkaos/gomakegen"></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-gomakegen-master"><img alt="codebeat badge" src="https://codebeat.co/badges/6f7a19c8-d78d-4062-a8cf-fdac4b8d1f85" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`gomakegen` is simple utility for generating makefiles for Golang applications.

### Installation

To build the `gomakegen` from scratch, make sure you have a working Go 1.14+ workspace (_[instructions](https://golang.org/doc/install)_), then:

```
go get github.com/essentialkaos/gomakegen
```

If you want to update `gomakegen` to latest stable release, do:

```
go get -u github.com/essentialkaos/gomakegen
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and OS X from [EK Apps Repository](https://apps.kaos.st/gomakegen/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) gomakegen
```

### Usage

```
Usage: gomakegen {options} dir

Options

  --glide, -g          Add target to fetching dependecies with glide
  --dep, -d            Add target to fetching dependecies with dep
  --mod, -m            Add target to fetching dependecies with go mod
  --metalinter, -M     Add target with metalinter check
  --strip, -S          Strip binaries
  --benchmark, -B      Add target to run benchmarks
  --verbose, -V        Enable verbose output for tests
  --race, -R           Add target to test race conditions
  --output, -o file    Output file (Makefile by default)
  --no-color, -nc      Disable colors in output
  --help, -h           Show this help message
  --version, -v        Show version

Examples

  gomakegen .
  Generate makefile for project in current directory and save as Makefile

  gomakegen $GOPATH/src/github.com/profile/project
  Generate makefile for github.com/profile/project and save as Makefile

  gomakegen $GOPATH/src/github.com/profile/project -o project.make
  Generate makefile for github.com/profile/project and save as project.make

```

### Build Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://github.com/essentialkaos/gomakegen/workflows/CI/badge.svg?branch=master)](https://github.com/essentialkaos/gomakegen/actions) |
| `develop` | [![CI](https://github.com/essentialkaos/gomakegen/workflows/CI/badge.svg?branch=develop)](https://github.com/essentialkaos/gomakegen/actions) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
