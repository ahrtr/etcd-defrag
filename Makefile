TESTFLAGS_RACE=-race=false
ifdef ENABLE_RACE
	TESTFLAGS_RACE=-race=true
endif

TESTFLAGS_CPU=
ifdef CPU
	TESTFLAGS_CPU=-cpu=$(CPU)
endif
TESTFLAGS = $(TESTFLAGS_RACE) $(TESTFLAGS_CPU) $(EXTRA_TESTFLAGS)

TESTFLAGS_TIMEOUT=10m
ifdef TIMEOUT
	TESTFLAGS_TIMEOUT=$(TIMEOUT)
endif

GOFILES = $(shell find . -name \*.go)

all: build

.PHONY: build
build:
	GO_BUILD_FLAGS="${GO_BUILD_FLAGS} -v -mod=readonly" ./build.sh

.PHONY: fmt
fmt:
	@echo "Verifying gofmt"
	@!(gofmt -l -s -d ${GOFILES} | grep '[a-z]')

	@echo "Verifying goimports"
	@!(go run golang.org/x/tools/cmd/goimports@latest -l -d ${GOFILES} | grep '[a-z]')

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -v ${TESTFLAGS} -timeout ${TESTFLAGS_TIMEOUT} ./...

clean:
	rm -rf ./bin
	rm -rf ./release
	rm -rf ./.idea
	rm -f etcd-defrag

