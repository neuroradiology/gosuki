.PHONY: all run deps docs build test debug

TARGET=gomark
# CGO_CFLAGS="-g -O2 -Wno-return-local-addr"
SRC := *.go
NVM_VERSIONS := $(HOME)/.config/nvm/versions/node
NVM_VERSION := $(shell cat ./web/.nvmrc)
export PATH := $(NVM_VERSIONS)/$(NVM_VERSION)/bin:$(PATH)
YARN := $(NVM_VERSIONS)/$(NVM_VERSION)/bin/yarn
DEBUG_FLAGS := -gcflags="all=-N -l"


#all: test build
all: build

# browser modules prototype
p_modules:
	@go run ./_prototype_modules/*

run: build
	@./$(TARGET)

debug: $(SRC)
	@#dlv debug . -- server
	@go build -v $(DEBUG_FLAGS) .

build: $(SRC)
	@echo building ...
	@# @CGO_CFLAGS=${CGO_CFLAGS} go build -o $(TARGET) *.go
	go build -v -o $(TARGET)


dev: build
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


test:
	@go test . ./...

testv:
	@go test -v . ./...
