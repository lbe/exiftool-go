.DEFAULT_GOAL := help

.PHONY: help build clean distclean test test-integration test-e2e test-all test-all-race cover fmt vet check golden-version golden-json golden-tree lint

help:
	@echo "Targets:"
	@echo "  build            go build -> ./bin/exiftool-go"
	@echo "  test             unit tests only (fast)"
	@echo "  test-integration integration tests (-tags integration)"
	@echo "  test-e2e         end-to-end tests (-tags e2e)"
	@echo "  test-all         unit + integration + e2e"
	@echo "  test-all-race    test-all with -race"
	@echo "  cover            coverage.out + coverage.html (all test tiers)"
	@echo "  fmt              go fmt ./..."
	@echo "  check            fmt, vet, test-all"
	@echo "  clean            remove ./bin"
	@echo "  distclean        same as clean"
	@echo "  golden-version   regen testdata/golden/version.stdout"
	@echo "  golden-json      regen testdata/golden/json_samsung_sm_g930p.json"
	@echo "  golden-tree      regen testdata/golden/tree_r_filename_sorted.stdout"
	@echo "  lint             golangci-lint run  --max-same-issues 0 ./..."

build:
	mkdir -p bin
	go build -v -o bin/exiftool-go ./cmd/exiftool-go

test:
	go test ./... -count=1 -cover

test-integration:
	go test -tags integration ./... -count=1

test-e2e:
	go test -tags e2e ./... -count=1

test-all:
	go test -tags "integration,e2e" ./... -count=1 -cover

test-all-race:
	go test -tags "integration,e2e" -race ./... -count=1

cover:
	go test -tags "integration,e2e" ./... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -n 1
	@echo "Coverage report: coverage.html"

fmt:
	go fmt ./...

lint:
	golangci-lint run  --max-same-issues 0 ./...

check: fmt vet test-all

clean:
	rm -rf bin

distclean: clean

golden-version:
	./scripts/exiftoolgen version

golden-json:
	./scripts/exiftoolgen json-samsung

golden-tree:
	./scripts/exiftoolgen tree-recurse
