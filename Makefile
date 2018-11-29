.PHONY: all run deps docs build test

TARGET=gomark


all: build

run: 
	@go run *.go

deps:
	go get

docs:
	@godoc2md gomark > API.md

build:
	@go build -o $(TARGET) *.go

test:
	@go test ./... | grep -v 'no test files'
