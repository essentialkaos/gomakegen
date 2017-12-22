################################################################################

# This Makefile generated by GoMakeGen 0.7.0 using next command:
# gomakegen --metalinter --strip .

################################################################################

.DEFAULT_GOAL := help
.PHONY = fmt all clean deps metalinter help

################################################################################

all: gomakegen ## Build all binaries

gomakegen: ## Build gomakegen binary
	go build -ldflags="-s -w" gomakegen.go

install: ## Install binaries
	cp gomakegen /usr/bin/gomakegen

uninstall: ## Uninstall binaries
	rm -f /usr/bin/gomakegen

deps: ## Download dependencies
	git config --global http.https://pkg.re.followRedirects true
	go get -d -v pkg.re/essentialkaos/ek.v9

fmt: ## Format source code with gofmt
	find . -name "*.go" -exec gofmt -s -w {} \;

metalinter: ## Install and run gometalinter
	test -s $(GOPATH)/bin/gometalinter || (go get -u github.com/alecthomas/gometalinter ; $(GOPATH)/bin/gometalinter --install)
	$(GOPATH)/bin/gometalinter --deadline 30s

clean: ## Remove generated files
	rm -f gomakegen

help: ## Show this info
	@echo -e '\nSupported targets:\n'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[33m%-12s\033[0m %s\n", $$1, $$2}'
	@echo -e ''

################################################################################
