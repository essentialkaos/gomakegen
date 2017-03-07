########################################################################################

.PHONY = fmt all clean deps

########################################################################################

all: gomakegen

gomakegen:
	go build gomakegen.go

deps:
	go get -v pkg.re/essentialkaos/ek.v7

fmt:
	find . -name "*.go" -exec gofmt -s -w {} \;

clean:
	rm -f gomakegen

########################################################################################

