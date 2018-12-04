.PHONY: all run deps docs build test

TARGET=gomark


all: test build

run: 
	@go run *.go

deps:
	go get

docs:
	@godoc2md gomark > API.md

build:
	@echo building ...
	@go build -o $(TARGET) *.go

test:
	@go test . ./...

testv:
	@go test -v . ./...
