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
	@go build -o $(TARGET) *.go

test:
	@go test . ./... | grep -v 'no test files'

testv:
	@go test -v . ./... | grep -v 'no test files'
