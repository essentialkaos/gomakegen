################################################################################

# This Makefile generated by GoMakeGen 2.0.0 using next command:
# gomakegen --strip .
#
# More info: https://kaos.sh/gomakegen

################################################################################

ifdef VERBOSE ## Print verbose information (Flag)
VERBOSE_FLAG = -v
endif

################################################################################

.DEFAULT_GOAL := help
.PHONY = fmt vet all clean deps update help

################################################################################

all: gomakegen ## Build all binaries

gomakegen: ## Build gomakegen binary
	go build $(VERBOSE_FLAG) -ldflags="-s -w" gomakegen.go

install: ## Install all binaries
	cp gomakegen /usr/bin/gomakegen

uninstall: ## Uninstall all binaries
	rm -f /usr/bin/gomakegen

deps: ## Download dependencies
	go get -d $(VERBOSE_FLAG) github.com/essentialkaos/ek

update: ## Update dependencies to the latest versions
	go get -d -u $(VERBOSE_FLAG) ./...

fmt: ## Format source code with gofmt
	find . -name "*.go" -exec gofmt -s -w {} \;

vet: ## Runs go vet over sources
	go vet -composites=false -printfuncs=LPrintf,TLPrintf,TPrintf,log.Debug,log.Info,log.Warn,log.Error,log.Critical,log.Print ./...

clean: ## Remove generated files
	rm -f gomakegen

help: ## Show this info
	@echo -e '\n\033[1mTargets:\033[0m\n'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[33m%-11s\033[0m %s\n", $$1, $$2}'
	@echo -e '\n\033[1mVariables:\033[0m\n'
	@grep -E '^ifdef [A-Z_]+ .*?## .*$$' $(abspath $(lastword $(MAKEFILE_LIST))) \
		| sed 's/ifdef //' \
		| awk 'BEGIN {FS = " .*?## "}; {printf "  \033[32m%-14s\033[0m %s\n", $$1, $$2}'
	@echo -e ''
	@echo -e '\033[90mGenerated by GoMakeGen 2.0.0\033[0m\n'

################################################################################
