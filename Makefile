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

make_ldflags = $(1) -X $(PKG)/pkg/build.Describe=$(VERSION)
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
	@ sed -i 's/LoggingMode = .*/LoggingMode = Dev/' pkg/logging/log.go
	$(GOBUILD) -tags "$(TAGS)" -o build/gosuki $(DEV_GCFLAGS) $(DEV_LDFLAGS) ./cmd/gosuki
	$(GOBUILD) -tags "$(TAGS)" -o build/suki $(DEV_GCFLAGS) $(DEV_LDFLAGS) ./cmd/suki


# debug: 
# 	@#dlv debug . -- server
# 	@go build -v $(DEV_GCFLAGS) -o build/gosuki ./cmd/gosuki

release: genimports
	@ sed -i 's/LoggingMode = .*/LoggingMode = Release/' pkg/logging/log.go
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
	@$(eval release_dir=gosuki-$(VERSION)-$(ARCH))
	@mkdir -p dist/$(release_dir)

	# Release package
	cp build/gosuki dist/$(release_dir)/
	cp build/suki dist/$(release_dir)/
	cp -r README.md LICENSE Makefile dist/$(release_dir)/
	tar -czf dist/$(release_dir).tar.gz -C dist/ $(release_dir)

	# Source Code ZIP
	@rm dist/$(release_dir)/{gosuki,suki}
	zip -r dist/gosuki-$(VERSION)-source.zip $$(git ls-files) -x .github\*
	zip -r dist/gosuki-$(VERSION)-source.zip mods/{mod-github-stars,reddit}
	@rm -rf dist/$(release_dir)

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
