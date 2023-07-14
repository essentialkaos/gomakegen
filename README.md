<p align="center"><a href="#readme"><img src="https://gh.kaos.st/gomakegen.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/gomakegen/ci"><img src="https://kaos.sh/w/gomakegen/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/gomakegen/codeql"><img src="https://kaos.sh/w/gomakegen/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="https://kaos.sh/r/gomakegen"><img src="https://kaos.sh/r/gomakegen.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/b/gomakegen"><img src="https://kaos.sh/b/6f7a19c8-d78d-4062-a8cf-fdac4b8d1f85.svg" alt="Codebeat badge" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`gomakegen` is simple utility for generating makefiles for Golang applications.

### Installation

To build the `gomakegen` from scratch, make sure you have a working Go 1.16+ workspace (_[instructions](https://go.dev/doc/install)_), then:

```
go install github.com/essentialkaos/gomakegen@latest
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and macOS from [EK Apps Repository](https://apps.kaos.st/gomakegen/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) gomakegen
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```bash
sudo gomakegen --completion=bash 1> /etc/bash_completion.d/gomakegen
```

ZSH:
```bash
sudo gomakegen --completion=zsh 1> /usr/share/zsh/site-functions/gomakegen
```

Fish:
```bash
sudo gomakegen --completion=fish 1> /usr/share/fish/vendor_completions.d/gomakegen.fish
```

### Man documentation

You can generate man page using next command:

```bash
gomakegen --generate-man | sudo gzip > /usr/share/man/man1/gomakegen.1.gz
```

### Usage

```
Usage: gomakegen {options} dir

Options

  --glide, -g          Add target to fetching dependencies with glide
  --dep, -d            Add target to fetching dependencies with dep
  --mod, -m            Add target to fetching dependencies with go mod (default for Go ≥ 1.18)
  --strip, -S          Strip binaries
  --benchmark, -B      Add target to run benchmarks
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

### CI Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/gomakegen/ci.svg?branch=master)](https://kaos.sh/w/gomakegen/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/gomakegen/ci.svg?branch=master)](https://kaos.sh/w/gomakegen/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
