BINS := gosuki suki

.PHONY: all run clean docs build test debug $(BINS) prepare

CGO_CFLAGS="-g -O2 -Wno-return-local-addr"
SRC := **/*.go
DEBUG_FLAGS := -gcflags="all=-N -l"
RELEASE_FLAGS := -ldflags="-s -w"
ifeq ($(origin TEST_FLAGS), environment)
	override TEST_FLAGS := $(TEST_FLAGS)
endif


all: prepare build

prepare:
	@mkdir -p build

build: $(BINS)

# run: gosuki
#  	@run command

$(BINS): 
	@echo building ... $@
	@# @CGO_CFLAGS=${CGO_CFLAGS} go build -o $@
	@go build -v -o build/$@ ./cmd/$@

debug: 
	@#dlv debug . -- server
	@go build -v $(DEBUG_FLAGS) -o build/gosuki ./cmd/gosuki

release:
	@echo building release ...
	go build -v $(RELEASE_FLAGS) -o build/gosuki ./cmd/gosuki

docs:
	@gomarkdoc -u ./... > docs/API.md

test:
	go test $(TEST_FLAGS) . ./...

clean:
	rm build/$(BINS)
