################################################################################

# This Makefile generated by GoMakeGen 3.0.2 using next command:
# gomakegen --mod --strip .
#
# More info: https://kaos.sh/gomakegen

################################################################################

ifdef VERBOSE ## Print verbose information (Flag)
VERBOSE_FLAG = -v
endif

COMPAT ?= 1.19
MAKEDIR = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
GITREV ?= $(shell test -s $(MAKEDIR)/.git && git rev-parse --short HEAD)

################################################################################

.DEFAULT_GOAL := help
.PHONY = fmt vet all install uninstall clean deps update init vendor mod-init mod-update mod-download mod-vendor help

################################################################################

all: gomakegen ## Build all binaries

gomakegen:
	@echo "[36;1mBuilding gomakegen…[0m"
	@go build $(VERBOSE_FLAG) -ldflags="-s -w -X main.gitrev=$(GITREV)" gomakegen.go

install: ## Install all binaries
	@echo "[36;1mInstalling binaries…[0m"
	@cp gomakegen /usr/bin/gomakegen

uninstall: ## Uninstall all binaries
	@echo "[36;1mRemoving installed binaries…[0m"
	@rm -f /usr/bin/gomakegen

init: mod-init ## Initialize new module

deps: mod-download ## Download dependencies

update: mod-update ## Update dependencies to the latest versions

vendor: mod-vendor ## Make vendored copy of dependencies

mod-init:
	@echo "[37m[1/2][0m [36;1mModules initialization…[0m"
ifdef MODULE_PATH ## Module path for initialization (String)
	@go mod init $(MODULE_PATH)
else
	@go mod init
endif

	@echo "[37m[2/2][0m [36;1mDependencies cleanup…[0m"
ifdef COMPAT ## Compatible Go version (String)
	@go mod tidy $(VERBOSE_FLAG) -compat=$(COMPAT) -go=$(COMPAT)
else
	@go mod tidy $(VERBOSE_FLAG)
endif

mod-update:
	@echo "[37m[1/4][0m [36;1mUpdating dependencies…[0m"
ifdef UPDATE_ALL ## Update all dependencies (Flag)
	@go get -u $(VERBOSE_FLAG) all
else
	@go get -u $(VERBOSE_FLAG) ./...
endif

	@echo "[37m[2/4][0m [36;1mStripping toolchain info…[0m"
	@grep -q 'toolchain ' go.mod && go mod edit -toolchain=none || :

	@echo "[37m[3/4][0m [36;1mDependencies cleanup…[0m"
ifdef COMPAT
	@go mod tidy $(VERBOSE_FLAG) -compat=$(COMPAT)
else
	@go mod tidy $(VERBOSE_FLAG)
endif

	@echo "[37m[4/4][0m [36;1mUpdating vendored dependencies…[0m"
	@test -d vendor && rm -rf vendor && go mod vendor $(VERBOSE_FLAG) || :

mod-download:
	@echo "[36;1mDownloading dependencies…[0m"
	@go mod download

mod-vendor:
	@echo "[36;1mVendoring dependencies…[0m"
	@rm -rf vendor && go mod vendor $(VERBOSE_FLAG) || :

fmt: ## Format source code with gofmt
	@echo "[36;1mFormatting sources…[0m"
	@find . -name "*.go" -exec gofmt -s -w {} \;

vet: ## Runs 'go vet' over sources
	@echo "[36;1mRunning 'go vet' over sources…[0m"
	@go vet -composites=false -printfuncs=LPrintf,TLPrintf,TPrintf,log.Debug,log.Info,log.Warn,log.Error,log.Critical,log.Print ./...

clean: ## Remove generated files
	@echo "[36;1mRemoving built binaries…[0m"
	@rm -f gomakegen

help: ## Show this info
	@echo -e '\n\033[1mTargets:\033[0m\n'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[33m%-9s\033[0m  %s\n", $$1, $$2}'
	@echo -e '\n\033[1mVariables:\033[0m\n'
	@grep -E '^ifdef [A-Z_]+ .*?## .*$$' $(abspath $(lastword $(MAKEFILE_LIST))) \
		| sed 's/ifdef //' \
		| awk 'BEGIN {FS = " .*?## "}; {printf "  \033[32m%-11s\033[0m  %s\n", $$1, $$2}'
	@echo -e ''
	@echo -e '\033[90mGenerated by GoMakeGen 3.0.2\033[0m\n'

################################################################################
