.PHONY: all run deps docs build test

TARGET=gomark
CGO_CFLAGS="-g -O2 -Wno-return-local-addr"


#all: test build
all: build

run: 
	@go run *.go

deps:
	go get

docs:
	@gomarkdoc -u ./... > API.md

build:
	@echo building ...
	@CGO_CFLAGS=${CGO_CFLAGS} go build -o $(TARGET) *.go

test:
	@go test . ./...

testv:
	@go test -v . ./...
