GO ?= go

.PHONY: build lint test check

build:
	$(GO) build ./...

lint:
	@out="$$(gofmt -l $$($(GO) list -f '{{.Dir}}' ./...))"; if [ -n "$$out" ]; then echo "gofmt: unformatted files:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...

test:
	$(GO) test -count=1 ./...

check: build lint test
