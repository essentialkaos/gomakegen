<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/r/gomakegen"><img src="https://kaos.sh/r/gomakegen.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/y/gomakegen"><img src="https://kaos.sh/y/55b2a258de5b4f0b9a75da00802f149d.svg" alt="Codacy badge" /></a>
  <a href="https://kaos.sh/w/gomakegen/ci"><img src="https://kaos.sh/w/gomakegen/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/gomakegen/codeql"><img src="https://kaos.sh/w/gomakegen/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`gomakegen` is simple utility for generating makefiles for Golang applications.

### Installation

To build the `gomakegen` from scratch, make sure you have a working Go 1.23+ workspace (_[instructions](https://go.dev/doc/install)_), then:

```
go install github.com/essentialkaos/gomakegen/v3@latest
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

<img src=".github/images/usage.svg"/>

### CI Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/gomakegen/ci.svg?branch=master)](https://kaos.sh/w/gomakegen/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/gomakegen/ci.svg?branch=master)](https://kaos.sh/w/gomakegen/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/.github/blob/master/CONTRIBUTING.md).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://kaos.dev"><img src="https://raw.githubusercontent.com/essentialkaos/.github/refs/heads/master/images/ekgh.svg"/></a></p>
