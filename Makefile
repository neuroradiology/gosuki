TARGET=gosuki
CMD_PATH := ./cmd/$(TARGET)

.PHONY: all run clean docs build test debug $(CMD_PATH)

# CGO_CFLAGS="-g -O2 -Wno-return-local-addr"
SRC := **/*.go
DEBUG_FLAGS := -gcflags="all=-N -l"
RELEASE_FLAGS := -ldflags="-s -w"


all: build

build: $(CMD_PATH)

$(CMD_PATH): $(SRC)
	@echo building ... $(CMD_PATH)
	@# @CGO_CFLAGS=${CGO_CFLAGS} go build -o $(TARGET) *.go
	@go build -v -o $(TARGET) $(CMD_PATH)

run: build
	@./$(TARGET)

debug: $(SRC)
	@#dlv debug . -- server
	@go build -v $(DEBUG_FLAGS) $

release: $(SRC)
	@echo building release ...
	go build -v $(RELEASE_FLAGS) -o $(TARGET)

docs:
	@gomarkdoc -u ./... > docs/API.md

test:
	@go test -v . ./...

clean:
	rm ./$(TARGET)
