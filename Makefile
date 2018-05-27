.PHONY: all run deps docs build

TARGET=gomark

all: build docs

run: 
	@go run *.go

deps:
	go get

docs:
	@godoc2md gomark > API.md

build:
	@go build -o $TARGET 

