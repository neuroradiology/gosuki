.PHONY: all run deps docs build test

TARGET=gomark
CGO_CFLAGS="-g -O2 -Wno-return-local-addr"
SRC := *.go
NVM_VERSIONS := $(HOME)/.config/nvm/versions/node
NVM_VERSION := $(shell cat ./web/.nvmrc)
export PATH := $(NVM_VERSIONS)/$(NVM_VERSION)/bin:$(PATH)
YARN := $(NVM_VERSIONS)/$(NVM_VERSION)/bin/yarn


#all: test build
all: build


run: build
	@./$(TARGET)

dev:
	@$(YARN) --cwd ./web develop &
	@caddy start
	@./$(TARGET) server
	@caddy stop

server:
	@caddy start
	@./$(TARGET) server
	@caddy stop

deps: caddy-dep
	go get

caddy-dep:
	@caddy version

docs:
	@gomarkdoc -u ./... > API.md

build:
	@echo building ...
	@CGO_CFLAGS=${CGO_CFLAGS} go build -o $(TARGET) *.go

test:
	@go test . ./...

testv:
	@go test -v . ./...
