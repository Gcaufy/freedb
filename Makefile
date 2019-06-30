GO             ?= cd golang && go
GOOS           ?= $(word 1, $(subst /, " ", $(word 4, $(shell go version))))

MAKEFILE       := $(realpath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR       := $(shell dirname $(MAKEFILE))
SOURCES        := $(wildcard *.go golang/*.go golang/*/*.go) $(MAKEFILE)

REVISION       := $(shell git log -n 1 --pretty=format:%h -- $(SOURCES))
BUILD_FLAGS    := -a -ldflags "-X main.revision=$(REVISION) -w -extldflags=$(LDFLAGS)" -tags "$(TAGS)"

BINARY32       := freedb-$(GOOS)_386
BINARY64       := freedb-$(GOOS)_amd64
BINARYARM5     := freedb-$(GOOS)_arm5
BINARYARM6     := freedb-$(GOOS)_arm6
BINARYARM7     := freedb-$(GOOS)_arm7
BINARYARM8     := freedb-$(GOOS)_arm8
BINARYPPC64LE  := freedb-$(GOOS)_ppc64le
VERSION        := $(shell awk -F= '/"version": / {print $1}' package.json | tr -d ":" | tr -d "\", :version")
RELEASE32      := freedb-$(VERSION)-$(GOOS)_386
RELEASE64      := freedb-$(VERSION)-$(GOOS)_amd64
RELEASEARM5    := freedb-$(VERSION)-$(GOOS)_arm5
RELEASEARM6    := freedb-$(VERSION)-$(GOOS)_arm6
RELEASEARM7    := freedb-$(VERSION)-$(GOOS)_arm7
RELEASEARM8    := freedb-$(VERSION)-$(GOOS)_arm8
RELEASEPPC64LE := freedb-$(VERSION)-$(GOOS)_ppc64le

# https://en.wikipedia.org/wiki/Uname
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	BINARY := $(BINARY64)
else ifeq ($(UNAME_M),amd64)
	BINARY := $(BINARY64)
else ifeq ($(UNAME_M),i686)
	BINARY := $(BINARY32)
else ifeq ($(UNAME_M),i386)
	BINARY := $(BINARY32)
else ifeq ($(UNAME_M),armv5l)
	BINARY := $(BINARYARM5)
else ifeq ($(UNAME_M),armv6l)
	BINARY := $(BINARYARM6)
else ifeq ($(UNAME_M),armv7l)
	BINARY := $(BINARYARM7)
else ifeq ($(UNAME_M),armv8l)
	BINARY := $(BINARYARM8)
else ifeq ($(UNAME_M),ppc64le)
	BINARY := $(BINARYPPC64LE)
else
$(error "Build on $(UNAME_M) is not supported, yet.")
endif


all: bin/$(BINARY)

bin:
	mkdir -p ../$@

ifeq ($(GOOS),windows)
release: bin/$(BINARY32) bin/$(BINARY64)
	cd bin && cp -f $(BINARY32) freedb.exe && zip $(RELEASE32).zip freedb.exe
	cd bin && cp -f $(BINARY64) freedb.exe && zip $(RELEASE64).zip freedb.exe
	cd bin && rm -f freedb.exe
else ifeq ($(GOOS),linux)
release: bin/$(BINARY32) bin/$(BINARY64) bin/$(BINARYARM5) bin/$(BINARYARM6) bin/$(BINARYARM7) bin/$(BINARYARM8) bin/$(BINARYPPC64LE)
	cd bin && cp -f $(BINARY32) freedb && tar -czf $(RELEASE32).tgz freedb
	cd bin && cp -f $(BINARY64) freedb && tar -czf $(RELEASE64).tgz freedb
	cd bin && cp -f $(BINARYARM5) freedb && tar -czf $(RELEASEARM5).tgz freedb
	cd bin && cp -f $(BINARYARM6) freedb && tar -czf $(RELEASEARM6).tgz freedb
	cd bin && cp -f $(BINARYARM7) freedb && tar -czf $(RELEASEARM7).tgz freedb
	cd bin && cp -f $(BINARYARM8) freedb && tar -czf $(RELEASEARM8).tgz freedb
	cd bin && cp -f $(BINARYPPC64LE) freedb && tar -czf $(RELEASEPPC64LE).tgz freedb
	cd bin && rm -f freedb
else
release: bin/$(BINARY32) bin/$(BINARY64)
	cd bin && cp -f $(BINARY32) freedb && tar -czf $(RELEASE32).tgz freedb
	cd bin && cp -f $(BINARY64) freedb && tar -czf $(RELEASE64).tgz freedb
	cd bin && rm -f freedb
endif

release-all: clean test
	GOOS=darwin  make release
	GOOS=linux   make release
	GOOS=freebsd make release
	GOOS=openbsd make release
	GOOS=windows make release

test: $(SOURCES)
	SHELL=/bin/sh GOOS= $(GO) test -v -tags "$(TAGS)" ./...

install: bin/freedb

clean:
	$(RM) -r ../bin

bin/$(BINARY32): $(SOURCES)
	GOARCH=386 $(GO) build $(BUILD_FLAGS) -o ../$@

bin/$(BINARY64): $(SOURCES)
	GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o ../$@

# https://github.com/golang/go/wiki/GoArm
bin/$(BINARYARM5): $(SOURCES)
	GOARCH=arm GOARM=5 $(GO) build $(BUILD_FLAGS) -o ../$@

bin/$(BINARYARM6): $(SOURCES)
	GOARCH=arm GOARM=6 $(GO) build $(BUILD_FLAGS) -o ../$@

bin/$(BINARYARM7): $(SOURCES)
	GOARCH=arm GOARM=7 $(GO) build $(BUILD_FLAGS) -o ../$@

bin/$(BINARYARM8): $(SOURCES)
	GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o ../$@

bin/$(BINARYPPC64LE): $(SOURCES)
	GOARCH=ppc64le $(GO) build $(BUILD_FLAGS) -o ../$@

bin/freedb: bin/$(BINARY) | bin
	cp -f bin/$(BINARY) bin/freedb

update:
	$(GO) get -u
	$(GO) mod tidy
