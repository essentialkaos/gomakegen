<p align="center"><a href="#readme"><img src="https://gh.kaos.st/gomakegen.svg"/></a></p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<p align="center">
  <a href="https://travis-ci.org/essentialkaos/gomakegen"><img src="https://travis-ci.org/essentialkaos/gomakegen.svg"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/gomakegen"><img src="https://goreportcard.com/badge/github.com/essentialkaos/gomakegen"></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-gomakegen-master"><img alt="codebeat badge" src="https://codebeat.co/badges/6f7a19c8-d78d-4062-a8cf-fdac4b8d1f85" /></a>
  <a href="https://essentialkaos.com/ekol"><img src="https://gh.kaos.st/ekol.svg"></a>
</p>

`gomakegen` is simple utility for generating makefiles for Golang applications.

### Installation

#### From source

Before the initial install, allow git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

To build the `gomakegen` from scratch, make sure you have a working Go 1.8+ workspace (_[instructions](https://golang.org/doc/install)_), then:

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
  --metalinter, -m     Add target with metalinter check
  --strip, -s          Strip binary
  --benchmark, -b      Add target to run benchmarks
  --verbose, -V        Enable verbose output for tests
  --race, -r           Add target to test race conditions
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
| `master` | [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=master)](https://travis-ci.org/essentialkaos/gomakegen) |
| `develop` | [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=develop)](https://travis-ci.org/essentialkaos/gomakegen) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[EKOL](https://essentialkaos.com/ekol)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
