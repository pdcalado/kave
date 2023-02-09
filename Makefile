VERSION ?= $(shell git describe --abbrev=7 || echo -n "unversioned")
VERSION_PACKAGE ?= github.com/pdcalado/kave/internal/version

LDFLAGS ?= "-X '$(VERSION_PACKAGE).Version=$(VERSION)' -s -w"

fmt:
	gofmt -w -s ./
	goimports -w  -local github.com/pdcalado/kave ./

lint:
	golangci-lint run -v

clean:
	rm -rf ./bin

bin/kave:
	go build -ldflags=$(LDFLAGS) -o bin/kave ./cmd/kave

bin/server:
	go build -ldflags=$(LDFLAGS) -o bin/server ./cmd/server

build: bin/kave bin/server

mocks/%:
	mkdir -p cmd/server/mocks
	mockgen -source=cmd/server/$* -destination=cmd/server/mocks/$*

mocks:
	$(MAKE) mocks/keyvalue_handler.go

test:
	go test -test.v ./...
