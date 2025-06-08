PKG := github.com/blob42/gosuki
CGO_CFLAGS="-g -Wno-return-local-addr"
SRC := **/*.go
GOBUILD := go build -v
GOINSTALL := go install -v
GOTEST := go test

# We only return the part inside the double quote here to avoid escape issues
# when calling the external release script. The second parameter can be used to
# add additional ldflags if needed (currently only used for the release).

VERSION := $(shell git describe --tags --dirty 2>/dev/null || echo "unknown")

make_ldflags = $(1) -X $(PKG)/build.Commit=$(VERSION)
#https://go.dev/doc/gdb
# disable gc optimizations
DEV_GCFLAGS := -gcflags "all=-N -l"
DEV_LDFLAGS := -ldflags "$(call make_ldflags)"

#TODO: add optimization flags
RELEASE_LDFLAGS := -ldflags "$(call make_ldflags, -s -w -buildid=)"

TAGS := linux
ifdef SYSTRAY
    TAGS += systray
endif


# TODO: remove, needed for testing mvsqlite
# SQLITE3_SHARED_TAGS := $(TAGS) libsqlite3

ifeq ($(origin TEST_FLAGS), environment)
	override TEST_FLAGS := $(TEST_FLAGS)
endif

# shared: TAGS = $(SQLITE3_SHARED_TAGS)


all: prepare build

prepare:
	@mkdir -p build

build: genimports
	@$(call print, "Building debug gosuki and suki.")
	$(GOBUILD) -tags "$(TAGS)" -o build/gosuki $(DEV_GCFLAGS) $(DEV_LDFLAGS) ./cmd/gosuki
	$(GOBUILD) -tags "$(TAGS)" -o build/suki $(DEV_GCFLAGS) $(DEV_LDFLAGS) ./cmd/suki

# run: gosuki
#  	@run command

# $(BINS): $(SRC)
# 	@echo building ... $@
# 	@# @CGO_CFLAGS=${CGO_CFLAGS} go build -o $@
# 	@go build -v -tags "$(TAGS)" -o build/$@ ./cmd/$@

# debug: 
# 	@#dlv debug . -- server
# 	@go build -v $(DEV_GCFLAGS) -o build/gosuki ./cmd/gosuki

release:
	@$(call print, "Building release gosuki and suki.")
	$(GOBUILD) -tags "$(TAGS)" -o build/gosuki $(RELEASE_LDFLAGS) ./cmd/gosuki
	$(GOBUILD) -tags "$(TAGS)" -o build/suki   $(RELEASE_LDFLAGS) ./cmd/suki

docs:
	@gomarkdoc -u ./... > docs/API.md


genimports: 
	@go run generate/imports.go | tee mods/generated_imports.go

# Distribution packaging
ARCH := x86_64

dist: clean release
	@mkdir -p dist/$(VERSION)-$(ARCH)
	@cp build/gosuki dist/$(VERSION)-$(ARCH)/
	@cp build/suki dist/$(VERSION)-$(ARCH)/
	@cp -r README.md LICENSE Makefile $(SRC) dist/$(VERSION)-$(ARCH)/
	@tar -czf dist/$(VERSION)-$(ARCH).tar.gz -C build/ .

	# create the source code zip
	@rm dist/$(VERSION)-$(ARCH)/{gosuki,suki}
	@cd dist/ && zip -r $(VERSION)-source.zip $(VERSION)-$(ARCH) && cd -
	@rm -rf dist/$(VERSION)-$(ARCH)

testsum:
ifeq (, $(shell which gotestsum))
	$(GOINSTALL) gotest.tools/gotestsum@latest
endif
	gotestsum -f dots-v2 $(TEST_FLAGS) . ./...

ci-test:
ifeq (, $(shell which gotestsum))
	$(GOINSTALL) gotest.tools/gotestsum@latest
endif
	gotestsum -f dots $(TEST_FLAGS) . ./...

clean:
	rm -rf build dist

.PHONY: \
 		all \
 		run \
 		clean \
 		docs \
 		build \
 		test \
 		testsum \
 		ci-test \
 		debug \
 		prepare \
 		shared \
		genimports \
 		dist
