.PHONY: run deps

run: 
	@go run *.go

deps:
	go get
