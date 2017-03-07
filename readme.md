# GoMakeGen [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=master)](https://travis-ci.org/essentialkaos/gomakegen) [![Go Report Card](https://goreportcard.com/badge/github.com/essentialkaos/gomakegen)](https://goreportcard.com/report/github.com/essentialkaos/gomakegen) [![License](https://gh.kaos.io/ekol.svg)](https://essentialkaos.com/ekol)

`gomakegen` is simple utility for generating makefiles for Golang applications.

* [Installation](#installation)
* [Usage](#usage)
* [Build Status](#build-status)
* [Contributing](#contributing)
* [License](#license)

## Installation

To build the `gomakegen` from scratch, make sure you have a working Go 1.6+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get github.com/essentialkaos/gomakegen
```

If you want update `gomakegen` to latest stable release, do:

```
go get -u github.com/essentialkaos/gomakegen
```

## Usage

```
Usage: gomakegen {options} dir

Options

  --output, -o       Output file (Makefile by default)
  --no-color, -nc    Disable colors in output
  --help, -h         Show this help message
  --version, -v      Show version

Examples

  gomakegen $GOPATH/src/github.com/profile/project
  Generate makefile for github.com/profile/project and save as Makefile

  gomakegen $GOPATH/src/github.com/profile/project -o project.make
  Generate makefile for github.com/profile/project and save as project.make

```

## Build Status

| Branch | Status |
|------------|--------|
| `master` | [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=master)](https://travis-ci.org/essentialkaos/gomakegen) |
| `develop` | [![Build Status](https://travis-ci.org/essentialkaos/gomakegen.svg?branch=develop)](https://travis-ci.org/essentialkaos/gomakegen) |

## Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

## License

[EKOL](https://essentialkaos.com/ekol)







