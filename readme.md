# GoMakeGen [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=master)](https://travis-ci.org/essentialkaos/gomakegen) [![Go Report Card](https://goreportcard.com/badge/github.com/essentialkaos/gomakegen)](https://goreportcard.com/report/github.com/essentialkaos/gomakegen) [![codebeat badge](https://codebeat.co/badges/6f7a19c8-d78d-4062-a8cf-fdac4b8d1f85)](https://codebeat.co/projects/github-com-essentialkaos-gomakegen-master) [![License](https://gh.kaos.io/ekol.svg)](https://essentialkaos.com/ekol)

`gomakegen` is simple utility for generating makefiles for Golang applications.

* [Installation](#installation)
* [Usage](#usage)
* [Build Status](#build-status)
* [Contributing](#contributing)
* [License](#license)

### Installation

#### From source

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)):

```
git config --global http.https://pkg.re.followRedirects true
```

To build the `gomakegen` from scratch, make sure you have a working Go 1.6+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get github.com/essentialkaos/gomakegen
```

If you want to update `gomakegen` to latest stable release, do:

```
go get -u github.com/essentialkaos/gomakegen
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and OS X from [EK Apps Repository](https://apps.kaos.io/gomakegen/latest).

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
  --output, -o file    Output file (Makefile by default)
  --no-color, -nc      Disable colors in output
  --help, -h           Show this help message
  --version, -v        Show version

Examples

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

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.io/ekgh.svg"/></a></p>
